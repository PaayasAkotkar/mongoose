mod misc;
mod mongoose_cutter;
mod mongoose_joiner;
use crate::mongoose_joiner::mio::join_chunks;
// use the created id after aaa/ and paste it next to it
fn main() {
    let bin = "./../../golang/tests\\bin\\greet\\aaa\\2f1ec57d_06f3_47c2_b188_489378c9a0d9";
    let dst = "./../../golang/tests/results";
    const MARK: &str = ".part_";
    join_chunks(bin, dst, MARK).expect("something went wrong");
    println!("suceeed joining chunks");
}
