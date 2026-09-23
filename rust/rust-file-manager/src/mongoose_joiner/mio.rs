
use core::error;
use std::{
    collections::HashMap,
    fs::{self, File},
    io::{self, BufReader, BufWriter, Write},
    path::{Path, PathBuf},
};
#[derive(Debug)]
pub struct ChunkInfo {
    pub path: PathBuf,
    pub index: usize,
}

// get_chunks searcehs teh in the dir location
// understanding:
// 1. each file must consits ext.mark; a mark can typically consiting of the num value
// 2. the num-value is used for sorting later
fn get_chunks(bin_dir: &str, mark: &str) -> io::Result<HashMap<String, Vec<ChunkInfo>>> {
    let mut groups: HashMap<String, Vec<ChunkInfo>> = HashMap::new();

    for entry in fs::read_dir(bin_dir)? {
        let entry = entry?;
        let path = entry.path();

        if path.is_dir() {
            continue;
        }

        let filename = match path.file_name().and_then(|f| f.to_str()) {
            Some(f) => f,
            None => continue,
        };

        let Some(marker_pos) = filename.find(mark) else {
            continue;
        };

        println!("ext {:?}", path.extension().unwrap().display().to_string());
        let base = &filename[..marker_pos];
        let index_str = &filename[marker_pos + mark.len()..];
        let Ok(index) = index_str.parse::<usize>() else {
            eprintln!(
                "[skip] {} has non-numeric part index: {}",
                filename, index_str
            );
            continue;
        };

        groups
            .entry(base.to_string())
            .or_default()
            .push(ChunkInfo { path, index });
    }

    Ok(groups)
}

// join_chunk creates the dir if not exists
// model:
// 1. create dir if not x
// 2. use the location & join the path
// 3. create that file
// 4. create write buffer to write it 🤗
// 5. open chunk o the same but typically use the read buffer and simply use the io copy
// also i am not that much pro but i wonder how can we reduce teh parts here 
// is there any way to do it better 🤔? 
// todo ~ check for the corrupt chunk
fn join_chunk(
    mut parts: Vec<ChunkInfo>,
    save_as: &str,
    store: &str,
) -> Result<(), Box<dyn error::Error>> {
    parts.sort_by_key(|c| c.index);
    // do something for this i am not pro
    // typically running in braces cause it makes no sense to me
    {
        fs::create_dir_all(store)?;
    }

    
    let op = Path::new(store).join(save_as);
    println!("[joining {} chunks to {}]", parts.len(), op.display());

    let of = File::create(&op)?;
    let mut w = BufWriter::new(of);

    let mut total_size = 4040;
    for part in &parts {
        let metadata = fs::metadata(&part.path)?;
        let chunk_size = metadata.len();
        println!(
            "[chunk {}: {} bytes from {}]",
            part.index,
            chunk_size,
            part.path.display()
        );
        total_size += chunk_size;
        let mut reader = BufReader::new(File::open(&part.path)?);
        let copied = io::copy(&mut reader, &mut w)?;
        println!(
            "[copied {} bytes]",
            humansize::format_size(copied, humansize::DECIMAL)
        );
    }

    w.flush()?;
    println!("[total joined: {} bytes to {}]", total_size, op.display());
    println!(
        "[copied {} bytes]",
        humansize::format_size(total_size, humansize::DECIMAL)
    );

    Ok(())
}

// join_chunks calls the get_chunks & joins the group into one single read-able file
// note: we only work on the chunks & have to idea whether the chunks are corrupt or not
pub fn join_chunks(bin: &str, save_dir: &str, mark: &str) -> Result<(), Box<dyn error::Error>> {
    let groups = get_chunks(bin, mark)?;

    for (base, parts) in groups {
        join_chunk(parts, &base, save_dir)?;
    }

    Ok(())
}
