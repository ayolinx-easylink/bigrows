# AGENTS.md

## Project Goal

Buat aplikasi CLI Golang untuk memproses file CSV besar.

Fitur utama:

1. Membaca semua file `.csv` dari folder tertentu.
2. Menampilkan daftar file CSV yang ditemukan.
3. User dapat memilih file yang ingin diproses.
4. User dapat input jumlah pembagian file.
5. File CSV diproses secara streaming agar hemat memory.
6. Output disimpan ke folder baru berdasarkan nama file.
7. Header CSV harus tetap ada di setiap file hasil split.

## Tech Stack

* Language: Go
* Minimum Go version: 1.22
* Standard library first.
* Hindari dependency eksternal kecuali benar-benar diperlukan.

## CLI Behavior

Aplikasi harus bisa dijalankan seperti ini:

```bash
go run main.go -dir ./data
```

Setelah dijalankan:

```text
CSV files found:
1. transaksi.csv
2. customer.csv
3. report.csv

Choose file number: 1
Split into how many parts: 5
```

Output:

```text
data/transaksi_split/
  transaksi_part_1.csv
  transaksi_part_2.csv
  transaksi_part_3.csv
  transaksi_part_4.csv
  transaksi_part_5.csv
```

## Functional Requirements

* Hanya baca file dengan ekstensi `.csv`.
* Abaikan folder dan file non-CSV.
* Jika folder tidak ditemukan, tampilkan error yang jelas.
* Jika tidak ada file CSV, tampilkan pesan yang jelas.
* Validasi input pilihan file.
* Validasi jumlah pembagian harus lebih dari 0.
* Gunakan `encoding/csv`.
* Gunakan streaming read/write, jangan load seluruh isi CSV ke memory.
* Gunakan `reader.FieldsPerRecord = -1` agar toleran terhadap jumlah kolom yang tidak konsisten.
* Header dari file sumber wajib ditulis ulang ke setiap file hasil split.
* Nama output folder:

  * `<nama_file>_split`
* Nama output file:

  * `<nama_file>_part_1.csv`
  * `<nama_file>_part_2.csv`
  * dst.

## Split Logic

Karena user memasukkan jumlah pembagian, aplikasi perlu:

1. Hitung total data rows tanpa header.
2. Hitung rows per part:

```text
rowsPerPart = ceil(totalRows / parts)
```

3. Baca ulang file dari awal.
4. Tulis data ke file part sesuai rowsPerPart.
5. Jangan membuat part kosong jika jumlah part lebih besar dari total rows.

## Code Structure

Gunakan struktur sederhana:

```text
.
├── AGENTS.md
├── go.mod
├── main.go
└── README.md
```

Jika kode mulai besar, boleh dipisah:

```text
.
├── cmd/
│   └── csvsplitter/
│       └── main.go
├── internal/
│   └── splitter/
│       ├── splitter.go
│       └── scanner.go
├── go.mod
├── README.md
└── AGENTS.md
```

## Coding Style

* Gunakan nama function yang jelas.
* Return error, jangan langsung `panic`.
* Gunakan `log.Fatal` hanya di entry point `main`.
* Pastikan file selalu ditutup dengan benar.
* Pastikan `writer.Flush()` dipanggil sebelum file ditutup.
* Setelah `Flush()`, cek `writer.Error()`.
* Jangan hardcode path.
* Jangan menimpa file tanpa sengaja jika bisa dihindari.
* Buat pesan CLI mudah dipahami.

## Suggested Functions

Minimal function yang disarankan:

```go
func listCSVFiles(dir string) ([]string, error)
func askFileChoice(files []string) (string, error)
func askParts() (int, error)
func countRows(filePath string) (totalRows int, header []string, err error)
func splitCSV(filePath string, parts int) error
func createOutputFile(outputDir, baseName string, partIndex int, header []string) (*os.File, *csv.Writer, error)
```

## Error Handling

Tangani kondisi berikut:

* Folder tidak ditemukan.
* Tidak ada file CSV.
* File tidak bisa dibuka.
* CSV kosong.
* CSV tidak punya header.
* Input pilihan bukan angka.
* Input pilihan di luar daftar.
* Input jumlah part kurang dari 1.
* Gagal membuat folder output.
* Gagal menulis file output.

## Testing

Buat minimal test untuk:

1. Folder kosong.
2. File CSV kecil dengan header.
3. Split menjadi 2 bagian.
4. Split dengan jumlah part lebih besar dari jumlah rows.
5. CSV dengan jumlah kolom tidak konsisten.

Gunakan temporary directory dari Go testing package.

## Performance Requirement

Aplikasi harus aman untuk CSV besar, termasuk file lebih dari 1 juta rows.

Dilarang:

* Membaca seluruh CSV ke slice besar.
* Menggunakan `ReadAll()`.
* Menggunakan pandas atau tool eksternal.
* Menyimpan semua rows di memory.

## README Requirement

Buat README.md berisi:

* Deskripsi aplikasi.
* Cara install.
* Cara run.
* Contoh struktur folder input.
* Contoh output.
* Catatan bahwa proses dilakukan secara streaming.
* Batasan aplikasi.

## Final Output Expected From Codex

Codex harus menghasilkan:

1. Source code Golang yang bisa dijalankan.
2. `go.mod`.
3. `README.md`.
4. Unit test jika memungkinkan.
5. Instruksi run yang jelas.

## Do Not Do

* Jangan buat aplikasi web.
* Jangan pakai database.
* Jangan pakai goroutine dulu kecuali benar-benar diperlukan.
* Jangan ubah isi data CSV.
* Jangan menghapus file input.
* Jangan overwrite file input asli.
