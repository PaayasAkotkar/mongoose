use core::error;
use human_format::{Formatter, Scales};
use humansize::ToF64;
use memmap2::Mmap;
use std::{
    cmp::{max, min},
    fs::{self, File},
    hash::{ Hasher},
    io::{BufWriter, Write},
    path::{Path, PathBuf},
    time::Duration,
};

use moka::sync::{Cache, CacheBuilder};

use zerocopy::{FromBytes, Immutable,  KnownLayout, little_endian::U64};

use crate::{
    mongoose_joiner::mio::join_chunks,
    misc::sizes::sizes::{GB, MB},
};

// todo- write your own formula for
// 1. data compressor
// 2. custom block chain mapping to track the fiel
#[allow(dead_code)]
fn a_c_e(data: &[u8], look_ahead: usize, w: usize) -> usize {
    let l = min(data.len(), w);
    let mut i = 1;
    let mut max_value = data[0];
    let mut max_pos = 0;

    while i < l {
        if data[i] < max_value {
            if i == (max_pos + look_ahead) {
                return i;
            }
        } else {
            max_value = data[i];
            max_pos = i;
        }
        i += 1;
    }
    return l;
}

struct AEParams {
    target_size: usize,  // expected chunk size // todo use this for allocation
    window: usize,       // w — passed as lookAhead to aec
    max_boundary: usize, // maxBoundary — passed as maxBoundary to aec
}

#[derive(FromBytes, Immutable, KnownLayout)]
#[allow(dead_code)]
struct Header {
    bytes: [U64; 1], // cap
    version: U64,
    file_size: U64,
}

// FindSizes returns the target for the computeAEP
fn find_sizes(cap: usize, file_size: usize, mut chunk: usize) -> usize {
    if chunk < 1 {
        chunk = 1
    }
    let mut size = file_size / chunk;
    if size > cap {
        size = cap;
    }
    return size as usize;
}

// compute_aec_params returns the params required for the AEC
fn compute_aec_params(fallback: usize, w_div: f64, mul: f64) -> AEParams {
    let target = max(fallback, 1);

    let w = target.to_f64() / w_div;

    let b = target.to_f64() * mul;

    return AEParams {
        target_size: target,
        window: w as usize,
        max_boundary: b as usize,
    };
}

fn normalize_path(p: &Path) -> PathBuf {
    let s = p.display().to_string();
    if let Some(stripped) = s.strip_prefix(r"\\?\") {
        PathBuf::from(stripped)
    } else {
        p.to_path_buf()
    }
}

#[derive(Debug, Clone)]
pub struct FileSliceInfo {
    pub sys_info: SysInfo,
    pub slice_info: SliceInfo,
}

#[derive(Debug, Clone)]
pub struct SliceInfo {
    pub ext: String,             // extension of the file we reading into & splitting
    pub mark_ext: String,        // extension that is wrote after the ext
    pub slices: Vec<Slice>,      // actually parts of the file
    pub stored_location: String, // location where the file parts are stored
    pub hash: String,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct Slice {
    #[allow(dead_code)]
    pub data: Vec<u8>,
}

#[allow(dead_code)]
#[derive(Debug, Clone)]
pub struct SysInfo {
    pub space_left: u64,
    pub total_space: u64,
    pub drive: String,
    pub src: String, // src-from which the requesst made
}

#[derive(Debug, Clone)]
pub struct Void {
    pub info: FileSliceInfo,
}
impl Void {
    pub fn get_sys_info(src: &str) -> Result<SysInfo, Box<dyn error::Error>> {
        let mut sinf = SysInfo {
            space_left: 0,
            total_space: 0,
            drive: String::from(""),
            src: src.to_string(),
        };
        let f_size = human_format::Formatter::new();
        let nt = sysinfo::Networks::new_with_refreshed_list();
        for (k, kd) in nt.iter() {
            for _k in kd.ip_networks() {
                println!("k {} d {}", k, _k.addr);
            }
        }
        let raw = fs::canonicalize(src).expect("file not found");
        let path = normalize_path(&raw);

        let ds = sysinfo::Disks::new_with_refreshed_list();
        for d in ds.list() {
            let m = d.mount_point();
            let n = normalize_path(m);
            let ts = d.total_space() / GB as u64; // not the accurate +3 gb
            let asp = d.available_space() / GB as u64; // not the accrate +1 gb or 300 mb
            sinf.drive = m.display().to_string();
            sinf.space_left = asp;
            sinf.total_space = ts;

            if path.starts_with(n) {
                println!(
                    "file save on drive {:?}
                     total space  {}
                     space left {}
                ",
                    m,
                    f_size.format(ts.to_f64()),
                    f_size.format(asp.to_f64()),
                );
            }
        }
        Ok((sinf))
    }

    fn _build_chunk(
        src: &str,
        store: &str,
        cuts: f64,
        w: usize,
    ) -> Result<FileSliceInfo, Box<dyn error::Error>> {
        match Self::create_groups(src) {
            Ok(fsi) => {
                let base_name = Path::new(src)
                    .file_name()
                    .and_then(|f| f.to_str())
                    .unwrap_or("output");

                if let Ok(fsi) = Self::save_groups(store, base_name, fsi) {
                    print!("suceed saving");
                    return Ok(fsi);
                } else {
                    return Err(format!("something went wrong").into());
                }
            }
            Err(e) => {
                eprintln!("{:?}", e);
                return Err(e);
            }
        }
    }

    pub fn build_chunk(
        src: &str,
        store: &str,
        cuts: f64,
        w: usize,
    ) -> Result<Self, Box<dyn error::Error>> {
        let cuts_clone = cuts.clone();
        let w_clone = w.clone();
        let cb: Cache<String, FileSliceInfo> = CacheBuilder::new(100 * MB)
            .time_to_live(Duration::from_secs(13))
            .time_to_idle(Duration::from_secs(24))
            .build();
        let key = format!("{}:{}:{}", src, store, w_clone);

        let fsi = cb.get_or_insert_with(key, move || {
            Void::_build_chunk(&src, &store, cuts_clone, w_clone).unwrap()
        });

        Ok(Void { info: fsi })
    }

    pub fn save_groups(
        store: &str,
        base_name: &str,
        mut inf: FileSliceInfo,
    ) -> Result<FileSliceInfo, Box<dyn error::Error>> {
        let mark = ".part_";
        fs::create_dir_all(store)?;
        for (i, g) in inf.slice_info.slices.iter().enumerate() {
            // formart:- file.pdf.part_n
            let path = Path::new(store).join(format!("{}{}{}", base_name, mark, i));
            let op = File::create(&path)?;
            let mut writer = BufWriter::new(op);
            writer.write_all(&g.data)?;
            writer.flush()?;
        }
        inf.slice_info.mark_ext = mark.to_string();
        inf.slice_info.stored_location = store.to_string();
        Ok(inf)
    }

    pub fn join_chunks(bin: &str, save_dir: &str, mark: &str) -> Result<(), Box<dyn error::Error>> {
        //match join_chunks(bin, save_dir, mark) {
        //    Ok(_) => {
        //        println!("suceed joining file");
        //    }
        //    Err(e) => {
        //        eprintln!("{:?}", e)
        //    }
        //}
        return join_chunks(bin, save_dir, mark);
    }

    pub fn create_groups(src: &str) -> Result<FileSliceInfo, Box<dyn error::Error>> {
        println!("working on dir {}", src);
        let fs = File::open(src)?;
        let md = fs.metadata()?;
        let formatter = Formatter::new();

        println!(
            "meta-data-size {} | header-size {}",
            formatter.format(md.len().to_f64()),
            formatter.format(std::mem::size_of::<Header>().to_f64()),
        );

        const CAP: usize = 1; // change to max file aceepted

        let mmap = unsafe { Mmap::map(&fs)? };
        let data: &[u8] = &mmap[..];

        let mut formatter = Formatter::new();
        formatter.with_scales(Scales::Binary()).with_units("B");
        println!("size {}", formatter.format(data.len().to_f64()));
        let mut g: Vec<Slice> = Vec::new();
        let file_cap = 30 * MB;
        let _cap = data.len();
        let cuts = 10;
        let w_div = 13.;
        //let mul = 210.;
        let mut df = rs_sha3_256::Sha3_256Hasher::default();

        if data.len() >= CAP {
            match Header::read_from_prefix(data) {
                Ok((_, rem)) => {
                    println!("bytes-len: {}", formatter.format(rem.len().to_f64()));
                    // Use full data instead of rem to avoid losing header bytes
                    g = Self::create_chunk(data, data, cuts, file_cap, w_div).unwrap();
                    let _ = df.write(&data.to_vec());
                }
                Err(e) => {
                    eprintln!("size {:?}", e)
                }
            }
        }
        let ex = Path::new(src).extension().expect("ext not found");
        let ext = ex.to_str().expect("extension not found");
        let bytes_result = rs_sha3_256::HasherContext::finish(&mut df);
        let fis = FileSliceInfo {
            slice_info: SliceInfo {
                slices: g,
                ext: String::from(ext),
                mark_ext: String::from(""),
                stored_location: String::from(""),
                hash: hex::encode(bytes_result),
            },

            sys_info: Self::get_sys_info(src).unwrap(),
        };

        println!("len {}", fis.slice_info.slices.len());
        Ok(fis)
    }

    pub fn create_chunk(
        data: &[u8],
        rem: &[u8],
        cuts: usize,
        cap_file_size: usize,
        w: f64,
    ) -> Result<Vec<Slice>, Box<dyn error::Error>> {
        let mut g: Vec<Slice> = Vec::new();
        let total = rem.len();
        let target = find_sizes(data.len(), cap_file_size, cuts);
        let mut cursor = 0;
        let p = compute_aec_params(target, w, cuts as f64);
        while total > cursor {
            let ch = &rem[cursor..total];
            let i = max(1, a_c_e(ch, p.window, p.max_boundary));
            let chunk = &rem[cursor..cursor + i];
            g.push(Slice {
                data: chunk.to_vec(),
            });
            cursor += i;
        }

        Ok(g)
    }
}
