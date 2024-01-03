# One Billion Row Challenge

A Go implementation of the [One Billion Row Challenge](https://github.com/gunnarmorling/1brc)

Read the accompanying [blog post](https://mrkaran.dev/posts/1brc/) for more details.

## Highlights

- Reads the file in chunks to efficiently to reduce I/O overhead.
- Spawns N workers for N cores for processing chunks.
- Mem allocation tweaks. Reuse byte buffers, avoid `strings.Split` for extra allocs etc
- Separate worker for aggregating results.

## Prerequisites

To generate the text file for these measurements, follow the steps outlined [here](https://github.com/gunnarmorling/1brc?tab=readme-ov-file#prerequisites).

After running the commands, I have a `measurements.txt` on my file system:

Example output after running the commands:

```sh
➜  1brc-go git:(main) du -sh measurements.txt
 13G	measurements.txt
➜  1brc-go git:(main) tail measurements.txt
Mek'ele;13.3
Kampala;50.8
Dikson;-3.7
Dodoma;20.3
San Diego;7.1
Chihuahua;20.3
Ngaoundéré;24.2
Toronto;12.7
Wrocław;12.6
Singapore;14.4
```

## Run the challenge

```sh
make run
```

## Results

Running the code on my laptop, which is Apple M2 Pro with 10‑core CPU, 32GB memory.

| Chunk Size | Time    |
| ---------- | ------- |
| 512.00 KB  | 23.756s |
| 1.00 MB    | 21.798s |
| 32.00 MB   | 19.501s |
| 16.00 MB   | 20.693s |
