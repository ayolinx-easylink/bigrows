# BigRows CSV Splitter

BigRows adalah aplikasi CLI Golang untuk memecah file CSV besar menjadi beberapa file lebih kecil. Proses baca dan tulis dilakukan secara streaming menggunakan `encoding/csv`, sehingga aman untuk file besar tanpa memuat seluruh isi CSV ke memory.

## Install

Pastikan Go 1.22 atau lebih baru sudah terpasang.

```bash
go mod tidy
go build -o bigrows
```

## Cara Run

Jalankan aplikasi langsung dengan lokasi file CSV dan jumlah pembagian:

```bash
go run main.go -file ./transaksi.csv -parts 5
```

Atau gunakan folder input agar aplikasi menampilkan daftar CSV dan Anda memilih file secara interaktif:

```bash
go run main.go -dir ./data
```

Jika `-parts` tidak diisi, aplikasi akan meminta jumlah pembagian lewat prompt.

Contoh interaksi:

```text
CSV files found:
1. customer.csv
2. report.csv
3. transaksi.csv

Choose file number: 3
Split into how many parts: 5
Split complete: 5 file(s) created in data/transaksi_split
```

Contoh command setelah build:

```bash
./bigrows -file ./data/transaksi.csv -parts 5
```

## Cara Menggunakan di Windows

Build aplikasi menjadi `csvsplitter.exe`:

```powershell
go build -o csvsplitter.exe
```

Jalankan langsung dengan lokasi file CSV:

```powershell
.\csvsplitter.exe -file .\dir\transaksi.csv -parts 2
```

Atau jalankan dengan lokasi folder, lalu pilih file dari daftar yang muncul:

```powershell
.\csvsplitter.exe -dir .\dir
```

Contoh interaksi:

```text
CSV files found:
1. customer.csv
2. report.csv
3. transaksi.csv

Choose file number: 3
Split into how many parts: 2
Split complete: 2 file(s) created in dir\transaksi_split
```

Jika file CSV berada di folder lain, gunakan path lengkap:

```powershell
.\csvsplitter.exe -file "D:\data\transaksi.csv" -parts 5
```

Output akan dibuat di folder yang sama dengan file CSV sumber:

```text
D:\data\transaksi_split\
  transaksi_part_1.csv
  transaksi_part_2.csv
  transaksi_part_3.csv
  transaksi_part_4.csv
  transaksi_part_5.csv
```

## Makefile

Beberapa command umum sudah tersedia:

```bash
make test
make build
make run-file INPUT_FILE=./dir/transaksi.csv PARTS=2
make run-dir ./dir
make clean
```

## Contoh Struktur Folder Input

```text
data/
  transaksi.csv
  customer.csv
  report.csv
  notes.txt
  archive/
```

Aplikasi hanya menampilkan file dengan ekstensi `.csv`. Folder dan file non-CSV diabaikan.

## Contoh Output

Jika file `data/transaksi.csv` dipecah menjadi 5 bagian, output dibuat di folder baru:

```text
data/transaksi_split/
  transaksi_part_1.csv
  transaksi_part_2.csv
  transaksi_part_3.csv
  transaksi_part_4.csv
  transaksi_part_5.csv
```

Header dari file sumber ditulis ulang ke setiap file hasil split.

## Catatan Streaming

Aplikasi melakukan dua kali pembacaan streaming:

1. Menghitung total baris data tanpa header.
2. Membaca ulang file dan menulis output per part.

Tidak ada penggunaan `ReadAll()` pada proses utama, dan data CSV tidak disimpan ke slice besar.

Aplikasi juga mengenali file ekspor yang membungkus seluruh baris dengan tanda kutip
dan menggandakan tanda kutip pada field. Format tersebut dinormalisasi per baris
sebelum diproses oleh `encoding/csv`.

Jika file payment record tidak memiliki header, aplikasi menambahkan header 19 kolom
payment record secara otomatis. Format tanpa header lainnya memakai nama
`column_1`, `column_2`, dan seterusnya. Hasil split selalu ditulis sebagai CSV standar.

## Batasan

- Output folder `<nama_file>_split` akan ditimpa jika sudah ada.
- File CSV tanpa header akan diberi header otomatis.
- File CSV tanpa baris data tidak akan dibuatkan output part kosong.
- Format CSV dibaca mengikuti perilaku standar `encoding/csv`.
- Format baris yang seluruhnya dibungkus tanda kutip tidak mendukung field dengan baris baru.
