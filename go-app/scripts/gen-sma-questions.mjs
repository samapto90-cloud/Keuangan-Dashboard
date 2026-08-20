/**
 * Generates 300 unique SMA questions into go-app/mmo/data/sma-edu-questions.json
 */
import { writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const now = Date.now();
const diffs = (i) => (i < 30 ? "EASY" : i < 60 ? "MEDIUM" : "HARD");

function item(id, subject, category, i, question, A, B, C, D, correct, explanation) {
  return {
    id,
    subject,
    category,
    grade: "SMA",
    difficulty: diffs(i),
    question,
    optionA: A,
    optionB: B,
    optionC: C,
    optionD: D,
    correctAnswer: correct,
    explanation,
    active: true,
    createdAt: now,
    updatedAt: now,
  };
}

function mathBank() {
  const out = [];
  let i = 0;
  const add = (q, A, B, C, D, ans, exp, cat) => {
    out.push(item(`math-${String(i + 1).padStart(3, "0")}`, "MATEMATIKA", cat, i, q, A, B, C, D, ans, exp));
    i++;
  };
  for (let n = 2; n <= 16; n++) {
    const a = n,
      b = n + 1,
      s = a + b;
    add(`${a} + ${b} = ?`, String(s - 1), String(s), String(s + 1), String(s + 2), "B", `${a} + ${b} = ${s}.`, "aljabar");
  }
  for (let n = 3; n <= 12; n++) {
    const a = n + 4,
      b = n,
      d = a - b;
    add(`${a} − ${b} = ?`, String(d - 1), String(d), String(d + 1), String(d + 2), "B", `${a} dikurangi ${b} adalah ${d}.`, "aljabar");
  }
  for (let n = 2; n <= 9; n++) {
    const p = n * (n + 1);
    add(`${n} × ${n + 1} = ?`, String(p - n), String(p), String(p + n), String(2 * n), "B", `${n} dikali ${n + 1} sama dengan ${p}.`, "aljabar");
  }
  add("Akar kuadrat dari 81 adalah …", "7", "8", "9", "10", "C", "9 × 9 = 81.", "aljabar");
  add("Akar kuadrat dari 144 adalah …", "10", "11", "12", "13", "C", "12 × 12 = 144.", "aljabar");
  add("Nilai dari 2³ adalah …", "6", "8", "9", "12", "B", "2 × 2 × 2 = 8.", "eksponen");
  add("Nilai dari 3² adalah …", "6", "8", "9", "12", "C", "3 × 3 = 9.", "eksponen");
  add("log₁₀ 100 = …", "1", "2", "10", "100", "B", "10² = 100 sehingga log₁₀ 100 = 2.", "logaritma");
  add("log₁₀ 1000 = …", "2", "3", "4", "10", "B", "10³ = 1000 sehingga hasilnya 3.", "logaritma");
  add("Jika f(x) = 2x + 1, maka f(3) = …", "5", "6", "7", "8", "C", "2(3)+1 = 7.", "fungsi");
  add("Jika f(x) = x², maka f(5) = …", "10", "20", "25", "30", "C", "5 × 5 = 25.", "fungsi");
  add("Penyelesaian 2x = 10 adalah x = …", "2", "4", "5", "8", "C", "x = 10/2 = 5.", "persamaan");
  add("Penyelesaian x + 7 = 15 adalah x = …", "6", "7", "8", "9", "C", "x = 15 − 7 = 8.", "persamaan");
  add("Himpunan x > 3 pada garis bilangan adalah …", "semua x kurang dari 3", "x = 3", "semua x lebih dari 3", "x = 0", "C", "Pertidaksamaan x > 3 memuat semua bilangan lebih besar dari 3.", "pertidaksamaan");
  add("Jumlah sudut segitiga adalah …", "90°", "180°", "270°", "360°", "B", "Jumlah sudut dalam segitiga selalu 180°.", "geometri");
  add("Keliling persegi sisi 6 cm adalah …", "12 cm", "18 cm", "24 cm", "36 cm", "C", "4 × 6 = 24 cm.", "geometri");
  add("Luas persegi sisi 5 cm adalah …", "10 cm²", "20 cm²", "25 cm²", "30 cm²", "C", "5 × 5 = 25 cm².", "geometri");
  add("Luas persegi panjang 8×3 adalah …", "11", "22", "24", "48", "C", "8 × 3 = 24.", "geometri");
  add("Pythagoras: jika a=3, b=4 maka c = …", "5", "6", "7", "12", "A", "3²+4²=9+16=25, √25=5.", "geometri");
  add("Mean dari 2, 4, 6 adalah …", "3", "4", "5", "6", "B", "(2+4+6)/3 = 4.", "statistika");
  add("Median dari 1, 3, 8 adalah …", "1", "3", "4", "8", "B", "Nilai tengah setelah diurutkan adalah 3.", "statistika");
  add("Peluang muncul angka genap pada dadu 1–6 adalah …", "1/6", "1/3", "1/2", "2/3", "C", "Ada 3 genap dari 6 sisi: 1/2.", "peluang");
  add("sin 90° = …", "0", "1/2", "1", "√2/2", "C", "sin 90° = 1.", "trigonometri");
  add("cos 0° = …", "0", "1/2", "1", "√3/2", "C", "cos 0° = 1.", "trigonometri");
  add("Suku ke-5 barisan 2,4,6,8,… adalah …", "8", "10", "12", "14", "B", "Barisan aritmetika beda 2; suku ke-5 = 10.", "barisan");
  add("Jumlah 1+2+3+4+5 = …", "10", "12", "15", "20", "C", "n(n+1)/2 = 5×6/2 = 15.", "deret");
  add("Determinan matriks [[1,0],[0,1]] adalah …", "0", "1", "2", "−1", "B", "Matriks identitas berdeterminannya 1.", "matriks");
  add("Jika 2x − 4 = 0 maka x = …", "1", "2", "4", "8", "B", "2x = 4, x = 2.", "persamaan");
  add("Nilai |−7| adalah …", "−7", "0", "7", "14", "C", "Nilai mutlak selalu nonnegatif: 7.", "aljabar");
  add("0,25 dalam pecahan adalah …", "1/5", "1/4", "1/3", "1/2", "B", "0,25 = 25/100 = 1/4.", "aljabar");
  add("50% dari 80 adalah …", "20", "30", "40", "50", "C", "0,5 × 80 = 40.", "aljabar");
  add("FPB 12 dan 18 adalah …", "3", "6", "9", "12", "B", "Faktor persekutuan terbesar adalah 6.", "aljabar");
  add("KPK 4 dan 6 adalah …", "8", "10", "12", "24", "C", "Kelipatan persekutuan terkecil adalah 12.", "aljabar");
  add("Jika an = 3n, suku ke-4 adalah …", "7", "9", "12", "15", "C", "3×4 = 12.", "barisan");
  add("Volume kubus rusuk 3 adalah …", "9", "18", "27", "36", "C", "3³ = 27.", "geometri");
  while (out.length < 75) {
    const k = out.length + 1;
    const x = k + 10;
    const y = k + 3;
    add(`${x} − ${y} = ?`, String(x - y - 1), String(x - y), String(x - y + 1), String(x + y), "B", `${x} − ${y} = ${x - y}.`, "aljabar");
  }
  return out.slice(0, 75);
}

const PAI = [
  ["Rukun Islam berjumlah …", "4", "5", "6", "7", "B", "Rukun Islam ada lima.", "akidah"],
  ["Rukun Iman berjumlah …", "5", "6", "7", "8", "B", "Rukun Iman ada enam.", "akidah"],
  ["Kitab suci umat Islam adalah …", "Injil", "Taurat", "Al-Qur'an", "Zabur", "C", "Al-Qur'an adalah kitab suci umat Islam.", "alquran"],
  ["Nabi dan rasul terakhir adalah …", "Nabi Musa", "Nabi Isa", "Nabi Muhammad", "Nabi Ibrahim", "C", "Nabi Muhammad SAW adalah penutup para nabi.", "akidah"],
  ["Arah kiblat saat salat adalah …", "Madinah", "Yerusalem", "Ka'bah di Mekah", "Kubah Shakhrah", "C", "Kiblat umat Islam adalah Ka'bah.", "fiqih"],
  ["Salat wajib dalam sehari semalam berjumlah …", "3", "4", "5", "7", "C", "Lima waktu: Subuh, Zuhur, Asar, Magrib, Isya.", "fiqih"],
  ["Puasa Ramadan termasuk …", "sunah", "wajib", "mubah", "makruh", "B", "Puasa Ramadan adalah rukun Islam yang wajib.", "fiqih"],
  ["Zakat fitrah ditunaikan pada …", "Idul Adha", "Ramadan/Idul Fitri", "Muharram", "Maulid", "B", "Zakat fitrah terkait Ramadan dan Idul Fitri.", "fiqih"],
  ["Haji dilaksanakan di …", "Madinah saja", "Yerusalem", "Mekah dan sekitarnya", "Kairo", "C", "Ibadah haji berpusat di Tanah Suci Mekah.", "fiqih"],
  ["Mengucapkan dua kalimat syahadat termasuk …", "sunah", "rukun Islam pertama", "hanya anjuran", "makruh", "B", "Syahadat adalah rukun Islam yang pertama.", "akidah"],
  ["Perilaku jujur dalam Islam disebut juga …", "riya", "amanah/siddiq", "hasad", "ghibah", "B", "Kejujuran (siddiq) dan amanah sangat ditekankan.", "akhlak"],
  ["Ghibah artinya …", "bersedekah", "menggunjing orang", "berpuasa", "berzikir", "B", "Ghibah adalah membicarakan aib orang lain.", "akhlak"],
  ["Menghormati orang tua dalam Islam termasuk …", "akhlak terpuji", "hal yang netral", "makruh", "syirik", "A", "Birrul walidain adalah akhlak mulia.", "akhlak"],
  ["Syirik berarti …", "menolong sesama", "menyekutukan Allah", "bersyukur", "berdoa", "B", "Syirik adalah menyekutukan Allah SWT.", "akidah"],
  ["Toleransi antarumat beragama dalam Islam didorong agar …", "mencampur akidah", "saling menghormati", "meninggalkan salat", "menghina agama lain", "B", "Islam mengajarkan hormat tanpa mencampur akidah.", "toleransi"],
  ["Sedekah berbeda dengan zakat karena sedekah …", "selalu wajib", "bersifat sunah/anjuran lebih luas", "hanya di Mekah", "mengganti salat", "B", "Zakat ada yang wajib; sedekah lebih luas dan sunah.", "muamalah"],
  ["Hari raya setelah Ramadan adalah …", "Idul Adha", "Idul Fitri", "Isra Mikraj", "Maulid", "B", "Idul Fitri menandai berakhirnya Ramadan.", "fiqih"],
  ["Hari raya kurban disebut …", "Idul Fitri", "Idul Adha", "Nuzulul Quran", "Tahun Baru Islam", "B", "Idul Adha terkait ibadah kurban.", "fiqih"],
  ["Membaca Al-Qur'an dianjurkan dengan …", "tergesa tanpa tartil", "tartil dan adab", "suara mengejek", "tanpa wudu jika mudah dihindari", "B", "Membaca Al-Qur'an dengan tartil dan adab.", "alquran"],
  ["Hadis adalah …", "kitab suci pengganti Al-Qur'an", "perkataan, perbuatan, dan ketetapan Nabi", "undang-undang negara", "syair Arab", "B", "Hadis meriwayatkan sunah Nabi SAW.", "hadis"],
  ["Wudu batal antara lain karena …", "tersenyum", "buang angin", "duduk", "berzikir", "B", "Hadats kecil seperti buang angin membatalkan wudu.", "fiqih"],
  ["Arah duduk tasyahud dalam salat menghadap …", "pintu masjid", "kiblat", "imam saja", "utara", "B", "Salat menghadap kiblat termasuk tasyahud.", "fiqih"],
  ["Puasa wajib ditinggalkan dengan sengaja tanpa uzur hukumnya …", "sunah", "berdosa/wajib qada", "mubah", "anjuran", "B", "Meninggalkan puasa Ramadan tanpa uzur adalah dosa dan wajib diqada.", "fiqih"],
  ["Zakat mal dikeluarkan dari …", "utang semata", "harta yang mencapai nisab", "hanya makanan", "hanya emas perhiasan anak", "B", "Zakat mal terkait harta mencapai nisab dan haul.", "fiqih"],
  ["Berbakti kepada ibu-bapak dalilnya sangat ditekankan setelah …", "olahraga", "tauhid/ibadah kepada Allah", "bisnis", "perang", "B", "Al-Qur'an menekankan tauhid lalu berbakti kepada orang tua.", "akhlak"],
  ["Menutup aurat dalam Islam termasuk …", "pilihan mode saja", "perintah syariat", "larangan", "adat lokal semata", "B", "Menutup aurat adalah perintah syariat.", "fiqih"],
  ["Masjidil Haram berada di …", "Madinah", "Mekah", "Yerusalem", "Istanbul", "B", "Masjidil Haram di Mekah.", "sejarah"],
  ["Masjid Nabawi berada di …", "Mekah", "Madinah", "Kufah", "Damaskus", "B", "Masjid Nabawi di Madinah.", "sejarah"],
  ["Hijrah Nabi SAW ke Madinah terjadi pada tahun …", "570 M", "622 M", "632 M", "610 M", "B", "Hijrah menjadi awal kalender Hijriah (622 M).", "sejarah"],
  ["Kota kelahiran Nabi Muhammad SAW adalah …", "Madinah", "Thaif", "Mekah", "Yaman", "C", "Nabi lahir di Mekah.", "sejarah"],
  ["Perilaku amanah berarti …", "khianat", "dapat dipercaya", "sombong", "iri", "B", "Amanah adalah dapat dipercaya.", "akhlak"],
  ["Hasad artinya …", "bersyukur", "dengki/iri", "dermawan", "sabar", "B", "Hasad adalah iri dengki.", "akhlak"],
  ["Sabar dalam Islam mencakup …", "hanya diam marah", "ketabahan dalam taat, musibah, dan menjauhi maksiat", "menunda salat", "membalas dendam", "B", "Sabar mencakup ketaatan dan menahan diri dari maksiat.", "akhlak"],
  ["Silaturahmi dianjurkan untuk …", "memutus keluarga", "mempererat hubungan kekerabatan", "riya", "menggunjing", "B", "Silaturahmi mempererat hubungan.", "akhlak"],
  ["Adab berbicara termasuk …", "berkata jujur dan tidak menyakiti", "bergegas menghina", "berbohong demi lucu", "membocorkan rahasia", "A", "Islam menuntun tutur kata yang baik.", "etika"],
  ["Pergaulan yang baik menurut PAI SMA menekankan …", "bebas tanpa batas", "menjaga kehormatan dan adab", "menghindari semua teman", "mencampur aduk aurat", "B", "Pergaulan dijaga dengan adab dan kehormatan.", "pergaulan"],
  ["Tanggung jawab murid di sekolah termasuk …", "menyontek", "belajar jujur dan menepati tugas", "membolos", "merusak fasilitas", "B", "Amanah belajar adalah tanggung jawab.", "tanggung jawab"],
  ["Niat dalam ibadah berfungsi …", "pengganti perbuatan", "membedakan ibadah dan adat serta keikhlasan", "agar dilihat orang", "menghapus rukun", "B", "Niat menentukan keikhlasan dan jenis amal.", "akidah"],
  ["Takwa secara ringkas adalah …", "kaya raya", "menjalankan perintah dan menjauhi larangan Allah", "terkenal", "menang debat", "B", "Takwa adalah taat kepada Allah.", "akidah"],
  ["Doa merupakan …", "pengganti usaha", "ibadah dan memohon kepada Allah disertai usaha", "sihir", "hal yang makruh", "B", "Doa adalah ibadah; tetap disertai ikhtiar.", "akidah"],
  ["Jenazah Muslim disalatkan dengan salat …", "Id", "Jenazah (gaib/hadir sesuai ketentuan)", "Istisqa", "Kusuf", "B", "Salat jenazah termasuk fardu kifayah.", "fiqih"],
  ["Makanan halal dan baik disebut …", "haram tayyib", "halalan tayyiban", "syubhat saja", "najis", "B", "Al-Qur'an menyebut halalan tayyiban.", "fiqih"],
  ["Riba dalam muamalah secara umum …", "dianjurkan", "diharamkan", "wajib", "sunah muakad", "B", "Riba diharamkan dalam Islam.", "muamalah"],
  ["Jual beli yang sah mensyaratkan …", "penipuan", "kerelaan dan objek yang halal/jelas", "paksaan", "barang najis tanpa keperluan syar'i", "B", "Akad jual beli butuh kerelaan dan objek yang sah.", "muamalah"],
  ["Menghormati guru termasuk …", "akhlak terpuji", "syirik", "bidah sesat otomatis", "makruh", "A", "Adab kepada guru adalah akhlak mulia.", "akhlak"],
  ["Membaca basmalah sebelum makan adalah …", "adab yang dianjurkan", "syirik", "wajib rukun iman", "larangan", "A", "Membaca basmalah termasuk adab makan.", "etika"],
  ["Qada salat berarti …", "mengganti salat yang tertinggal", "membatalkan wudu", "zakat", "haji", "A", "Qada adalah mengganti ibadah yang tertinggal.", "fiqih"],
  ["Nisab zakat terkait …", "batas minimal harta terkena zakat", "jumlah rakaat", "arah kiblat", "nama surat", "A", "Nisab adalah ambang harta wajib zakat.", "fiqih"],
  ["Lailatul Qadar terjadi pada …", "puluh hari terakhir Ramadan (malam ganjil, menurut riwayat)", "Idul Adha saja", "setiap Senin", "Muharram wajib", "A", "Lailatul Qadar dicari di akhir Ramadan.", "alquran"],
  ["Surat Al-Fatihah dibaca dalam …", "setiap rakaat salat", "hanya Jumat", "hanya jenazah", "hanya Id", "A", "Al-Fatihah adalah rukun salat pada tiap rakaat.", "fiqih"],
  ["Berwudu sebelum menyentuh mushaf menurut banyak ulama …", "dianjurkan/diwajibkan sesuai pendapat", "diharamkan", "menghapus iman", "mengganti salat", "A", "Menyentuh mushaf umumnya disertai kesucian/wudu.", "fiqih"],
  ["Nama malaikat yang menyampaikan wahyu adalah …", "Mikail", "Jibril", "Israfil", "Izrail", "B", "Jibril menyampaikan wahyu kepada para nabi.", "akidah"],
  ["Iman kepada hari akhir termasuk …", "rukun iman", "rukun Islam", "sunah biasa", "adat Mekah", "A", "Hari akhir adalah salah satu rukun iman.", "akidah"],
  ["Berbuat baik kepada tetangga termasuk …", "ajaran Islam", "hal yang dilarang", "hanya budaya modern", "syirik", "A", "Nabi menekankan hak tetangga.", "akhlak"],
  ["Menjaga lisan dari dusta termasuk …", "akhlak", "riya wajib", "ghibah sunah", "hasad", "A", "Menjaga lisan adalah bagian akhlak.", "akhlak"],
  ["Salat Jumat wajib bagi …", "muslim mukallaf laki-laki yang memenuhi syarat", "semua anak kecil", "nonmuslim", "hanya imam", "A", "Jumat wajib bagi yang memenuhi syarat syar'i.", "fiqih"],
  ["Azan berfungsi …", "pemberitahuan masuk waktu salat", "pengganti salat", "zakat", "haji", "A", "Azan menyeru masuknya waktu salat.", "fiqih"],
  ["Niat puasa Ramadan pada malam hari menurut banyak ulama …", "diperlukan untuk puasa wajib", "diharamkan", "mengganti sahur", "hanya untuk sunah", "A", "Niat adalah bagian penting puasa wajib.", "fiqih"],
  ["Sahur dianjurkan karena …", "mengikuti sunah dan menolong puasa", "membatalkan puasa", "wajib zakat", "mengganti magrib", "A", "Sahur adalah sunah yang membantu puasa.", "fiqih"],
  ["Berbuka puasa saat magrib adalah …", "waktu berbuka yang disyariatkan", "makruh", "wajib ditunda ke isya", "haram", "A", "Berbuka dilakukan setelah masuk magrib.", "fiqih"],
  ["Ukhuwah Islamiyah berarti …", "persaudaraan sesama muslim", "permusuhan", "riba", "syirik", "A", "Ukhuwah adalah persaudaraan dalam Islam.", "akhlak"],
  ["Musyawarah dalam Islam disebut juga …", "syura", "riba", "ghibah", "hasad", "A", "Syura adalah musyawarah.", "muamalah"],
  ["Menepati janji termasuk …", "amanah", "khianat", "nifak terpuji", "makruh", "A", "Menepati janji adalah sifat amanah; ingkar janji tercela.", "akhlak"],
  ["Nifak adalah …", "kemunafikan", "kejujuran", "zakat", "haji", "A", "Nifak berarti munafik.", "akidah"],
  ["Bersyukur atas nikmat Allah diwujudkan dengan …", "taat dan menggunakan nikmat secara benar", "kufur", "menyombongkan diri", "menghamun", "A", "Syukur lahir dalam hati, lisan, dan perbuatan.", "akhlak"],
  ["Isra Mikraj berkaitan dengan …", "perjalanan Nabi SAW dan perintah salat", "zakat fitrah", "idul adha saja", "puasa Syawal wajib", "A", "Isra Mikraj terkait peristiwa Nabi dan salat wajib.", "sejarah"],
  ["Piagam Madinah penting karena …", "mengatur kehidupan bersama di Madinah", "menghapus salat", "mewajibkan riba", "melarang zakat", "A", "Piagam Madinah adalah dokumen tata hidup bersama.", "sejarah"],
  ["Khulafaur Rasyidin berjumlah …", "3", "4", "5", "12", "B", "Abu Bakar, Umar, Usman, Ali.", "sejarah"],
  ["Membaca hamdalah setelah bersin yang baik adalah adab …", "Islam", "yang dilarang", "syirik", "makruh tahrim", "A", "Mengucapkan hamdalah termasuk adab bersin.", "etika"],
  ["Menjaga kebersihan termasuk bagian dari …", "iman/akhlak Islam", "hal yang diharamkan", "syirik", "bidah otomatis sesat", "A", "Kebersihan dinilai bagian iman.", "akhlak"],
  ["Berdoa untuk kedua orang tua setelah mereka wafat …", "dianjurkan", "diharamkan mutlak", "mengganti tauhid", "syirik wajib", "A", "Doa anak saleh bermanfaat bagi orang tua.", "akhlak"],
  ["Menuntut ilmu dalam Islam …", "dianjurkan/diwajibkan sesuai konteks", "diharamkan bagi perempuan secara mutlak", "hanya untuk imam", "makruh", "A", "Menuntut ilmu sangat ditekankan.", "akhlak"],
  ["Tidak berlebihan (israf) dalam belanja termasuk …", "ajaran sederhana", "perintah riba", "wajib boros", "syirik", "A", "Islam melarang berlebih-lebihan.", "muamalah"],
  ["Menghormati rumah ibadah agama lain termasuk …", "sikap toleran yang sopan", "mencampur akidah wajib", "murtad otomatis", "haram silaturahmi", "A", "Sopan santun publik tidak sama dengan mencampur akidah.", "toleransi"],
  ["Niat yang ikhlas berarti …", "beribadah karena Allah bukan pamer", "riya", "sum'ah wajib", "ghibah", "A", "Ikhlas lawan dari riya.", "akidah"],
];

function paiBank() {
  return PAI.slice(0, 75).map((row, i) =>
    item(`pai-${String(i + 1).padStart(3, "0")}`, "PAI", row[7], i, row[0], row[1], row[2], row[3], row[4], row[5], row[6]),
  );
}

const ENG = [
  ["The past tense of 'go' is …", "goed", "gone", "went", "going", "C", "The simple past of go is went.", "tenses"],
  ["She ____ to school every day.", "go", "goes", "going", "gone", "B", "Third person singular takes goes.", "tenses"],
  ["'Big' is the opposite of …", "large", "huge", "small", "tall", "C", "Small is the common antonym of big.", "vocabulary"],
  ["A word that names a person, place, or thing is a …", "verb", "adjective", "noun", "adverb", "C", "Nouns name people, places, things.", "grammar"],
  ["The plural of 'child' is …", "childs", "children", "childes", "child", "B", "Child → children (irregular plural).", "grammar"],
  ["I have ____ apple.", "a", "an", "the a", "some a", "B", "Use an before a vowel sound.", "grammar"],
  ["They ____ playing football now.", "is", "am", "are", "be", "C", "They takes are + V-ing.", "tenses"],
  ["Passive: 'They built the house' → The house ____.", "built", "was built", "is build", "were build", "B", "Past passive: was/were + past participle.", "passive"],
  ["If it rains, we ____ at home.", "stays", "will stay", "stayed", "staying", "B", "First conditional: If + present, will + V1.", "conditional"],
  ["He said, 'I am tired.' → He said that he ____ tired.", "is", "was", "were", "be", "B", "Reported speech backshifts am → was.", "reported"],
  ["A text that tells how to do something is a …", "narrative", "procedure", "recount only", "anecdote only", "B", "Procedure text explains steps.", "procedure"],
  ["Narrative text usually has …", "steps to cook", "orientation–complication–resolution", "only graphs", "a recipe list only", "B", "Narratives typically have those stages.", "narrative"],
  ["Descriptive text mainly …", "argues a policy", "describes a person/place/thing", "retells a legend only", "gives commands only", "B", "Description portrays features.", "descriptive"],
  ["Analytical exposition aims to …", "tell a fairy tale", "persuade with arguments", "teach cooking steps", "list prices only", "B", "It presents a thesis and arguments.", "analytical"],
  ["'However' is used to show …", "addition only", "contrast", "time only", "place only", "B", "However signals contrast.", "grammar"],
  ["Choose the correct sentence.", "She don't like tea.", "She doesn't like tea.", "She not like tea.", "She no likes tea.", "B", "Doesn't + base verb for she.", "grammar"],
  ["The synonym of 'happy' is …", "sad", "angry", "glad", "tired", "C", "Glad means happy.", "vocabulary"],
  ["The antonym of 'hot' is …", "warm", "cold", "boiling", "spicy", "B", "Cold is the common opposite of hot.", "vocabulary"],
  ["I ____ my homework yesterday.", "do", "did", "does", "doing", "B", "Yesterday requires past tense did.", "tenses"],
  ["We have lived here ____ 2018.", "for", "since", "during", "while", "B", "Since + starting point.", "tenses"],
  ["There ____ a book on the table.", "am", "is", "are", "be", "B", "A book is singular → is.", "grammar"],
  ["This is ____ interesting story.", "a", "an", "the a", "any a", "B", "Interesting begins with a vowel sound → an.", "grammar"],
  ["Comparative of 'tall' is …", "tallest", "more tall", "taller", "most tall", "C", "Short adjectives take -er.", "grammar"],
  ["Superlative of 'good' is …", "goodest", "better", "best", "more good", "C", "Good → better → best.", "grammar"],
  ["A letter of invitation is a …", "narrative", "functional text", "poem only", "novel", "B", "Invitations are short functional texts.", "functional"],
  ["'Please turn off the lamp' is a …", "narrative complication", "request/command", "thesis statement", "orientation only", "B", "It is an imperative request.", "functional"],
  ["Which is a question tag for 'You are ready'?", "aren't they?", "aren't you?", "isn't you?", "don't you?", "B", "You are → aren't you?", "grammar"],
  ["The gerund of 'read' is …", "readed", "reading", "to readed", "reads", "B", "V-ing form reading can be a gerund.", "grammar"],
  ["I look forward to ____ you.", "meet", "meeting", "met", "meets", "B", "After to (preposition) use gerund.", "grammar"],
  ["Neither of the answers ____ correct.", "are", "is", "be", "were always", "B", "Neither is singular → is.", "grammar"],
  ["The man ____ is talking is my uncle.", "which", "who", "where", "when", "B", "Who refers to a person.", "grammar"],
  ["This is the school ____ I studied.", "who", "which", "where", "whose", "C", "Where for place.", "grammar"],
  ["How ____ water do you need?", "many", "much", "a few", "several", "B", "Water is uncountable → much.", "grammar"],
  ["How ____ apples are there?", "much", "many", "little", "amount", "B", "Apples are countable → many.", "grammar"],
  ["She can ____ the piano.", "plays", "play", "playing", "played", "B", "Modal + base verb.", "grammar"],
  ["You must not ____ here.", "smoking", "smoke", "to smoke", "smokes", "B", "Must not + base verb.", "grammar"],
  ["I would rather ____ tea than coffee.", "drink", "drinking", "drank", "to drank", "A", "Would rather + base verb.", "grammar"],
  ["The book was written ____ a famous author.", "from", "by", "with", "at", "B", "Passive agent uses by.", "passive"],
  ["If I ____ rich, I would travel.", "am", "were", "was being", "be", "B", "Second conditional: If + past, would.", "conditional"],
  ["She asked me where I ____.", "live", "lived", "living", "lives", "B", "Reported question backshifts.", "reported"],
  ["A recount text retells …", "future plans only", "past events", "product specs only", "lab tools only", "B", "Recounts past experience.", "reading"],
  ["The main idea of a paragraph is the …", "topic sentence idea", "comma", "page number", "font", "A", "Main idea is often in the topic sentence.", "reading"],
  ["'In addition' shows …", "contrast", "addition", "result only", "place", "B", "In addition adds information.", "grammar"],
  ["'Therefore' shows …", "contrast", "result/conclusion", "example only", "time only", "B", "Therefore introduces a result.", "grammar"],
  ["Choose correct spelling.", "recieve", "receive", "receve", "receeve", "B", "i before e except after c: receive.", "vocabulary"],
  ["An essay should have a …", "thesis", "only emojis", "no paragraphs", "random list only", "A", "Academic essays need a thesis.", "writing"],
  ["'Their' refers to …", "place", "possession", "time", "weather", "B", "Their is a possessive determiner.", "vocabulary"],
  ["'There' often refers to …", "possession", "place/existence", "person", "past tense", "B", "There can mark place or existence.", "vocabulary"],
  ["I ____ TV when you called.", "watch", "was watching", "watches", "am watch", "B", "Past continuous for interrupted action.", "tenses"],
  ["By next year, she ____ graduated.", "has", "will have", "have", "had have", "B", "Future perfect: will have + V3.", "tenses"],
  ["The news ____ interesting.", "are", "is", "were always", "be", "B", "News is singular in agreement.", "grammar"],
  ["Scissors ____ on the table.", "is", "are", "was", "be", "B", "Scissors take plural verb.", "grammar"],
  ["He is interested ____ music.", "on", "in", "at", "to", "B", "Interested in + noun.", "grammar"],
  ["She is good ____ mathematics.", "in", "at", "on", "to", "B", "Good at + skill.", "grammar"],
  ["Don't forget ____ the door.", "locking", "to lock", "lock to", "locked", "B", "Forget to do (future duty).", "grammar"],
  ["I stopped ____ because I was tired.", "to run", "running", "run", "ran", "B", "Stop + gerund = quit the activity.", "grammar"],
  ["A caption usually …", "explains an image briefly", "is a novel chapter", "replaces grammar", "is a dictionary", "A", "Captions label pictures.", "functional"],
  ["An announcement is meant to …", "hide information", "inform people of news/events", "write a legend", "teach algebra", "B", "Announcements inform.", "functional"],
  ["'Could you help me?' is …", "an offer only", "a polite request", "a threat", "a proverb", "B", "Could you… is a polite request.", "functional"],
  ["The moral of a fable is the …", "setting", "lesson", "page", "author age", "B", "Fables teach a moral.", "narrative"],
  ["Skimming means reading …", "every word slowly", "quickly for gist", "only footnotes", "backwards", "B", "Skim for general idea.", "reading"],
  ["Scanning means reading to find …", "a specific detail", "the whole novel theme only", "rhyme scheme only", "paper size", "A", "Scan for particular facts.", "reading"],
  ["A topic sentence is usually …", "in the index", "the sentence that states the paragraph topic", "a footnote", "the title of a book only", "B", "It introduces the paragraph topic.", "reading"],
  ["'Although it rained, we went out.' Although shows …", "result", "contrast", "addition", "time only", "B", "Although introduces contrast.", "grammar"],
  ["The correct article: ____ honest man", "a", "an", "the a", "no never", "B", "Honest starts with a vowel sound.", "grammar"],
  ["She has already ____ the letter.", "write", "written", "wrote", "writing", "B", "Present perfect: has + past participle.", "tenses"],
  ["Neither John nor his friends ____ coming.", "is", "are", "be", "was always", "B", "Agreement with the nearer noun friends.", "grammar"],
  ["The teacher told the students ____ quiet.", "be", "to be", "being", "been", "B", "Tell someone to + V.", "grammar"],
  ["Which sentence is simple present?", "She is eating now.", "She eats rice.", "She ate rice.", "She will eat.", "B", "Eats is simple present.", "tenses"],
  ["A procedure's language feature often uses …", "imperatives", "only past perfect", "only metaphors", "no verbs", "A", "Steps use commands: Cut, Mix, Heat.", "procedure"],
  ["'First, then, finally' are …", "temporal conjunctions", "nouns", "pronouns only", "articles", "A", "They sequence steps.", "procedure"],
  ["An argument in exposition should be …", "unsupported opinion only", "supported with reasons/evidence", "a recipe", "a caption only", "B", "Arguments need support.", "analytical"],
  ["The word 'rapidly' is an …", "noun", "adverb", "pronoun", "article", "B", "-ly manner word is an adverb.", "grammar"],
  ["Choose the countable noun.", "water", "rice", "apple", "advice", "C", "Apple can be counted; the others are typically uncountable.", "grammar"],
  ["The passive of 'Someone stole my bag' is …", "My bag stole.", "My bag was stolen.", "My bag is steal.", "My bag stolen.", "B", "Past passive: was stolen.", "passive"],
];

function engBank() {
  return ENG.slice(0, 75).map((row, i) =>
    item(`eng-${String(i + 1).padStart(3, "0")}`, "BAHASA_INGGRIS", row[7], i, row[0], row[1], row[2], row[3], row[4], row[5], row[6]),
  );
}

const JAWA = [
  ["Unggah-ungguh basa tegese …", "aturan busana", "tata krama basa miturut papan lan lawan guneman", "jenis tembang", "aksara Latin", "B", "Unggah-ungguh ngatur undha-usuk basa.", "unggah-ungguh"],
  ["Basa ngoko kanggo …", "guneman karo wong kang dihormati banget tanpa alesan", "guneman akrab/sepadha utawa luwih enom miturut konteks", "mung kanggo raja", "mung aksara Jawa", "B", "Ngoko kanggo sesrawungan akrab utawa sepadha, miturut unggah-ungguh.", "ngoko"],
  ["Basa krama kanggo …", "ngina liyan", "ngurmati lawan guneman", "dolanan bocah cilik mung", "nulis angka", "B", "Krama minangka undha-usuk kang luwih ngormati.", "krama"],
  ["Krama inggil kanggo …", "nyendhu awak dhewe", "ngurmati wong liya (awaké wong kang dihormati)", "ngoko kasar", "basa Walanda", "B", "Krama inggil kanggo ngurmati liyan, dudu kanggo awak dhewe.", "krama inggil"],
  ["Tembung 'nedha' iku krama inggil saka …", "turu", "mangan", "ngombe", "lunga", "B", "Mangan → dhahar/nedha (krama inggil).", "kosakata"],
  ["Tembung 'sare' tegese …", "mangan", "turu", "ngombe", "mlaku", "B", "Sare = turu (krama inggil).", "kosakata"],
  ["Tembung 'tindak' krama inggil saka …", "lunga", "lungguh", "ngadeg", "lungguh lesehan", "A", "Lunga → tindak.", "kosakata"],
  ["Tembung 'pinarak' nduweni teges …", "mangan", "lungguh / pinarak", "turu", "ngombe", "B", "Pinarak gegayutan lungguh kanthi hormat.", "kosakata"],
  ["Paribasan iku …", "ukara kiasan tetembungan tetep", "tembang macapat", "aksara swara", "candra sengkala mung", "A", "Paribasan: ungkapan kiasan tetembungan.", "paribasan"],
  ["Bebasan iku …", "pepindhan kahanan kanthi kiasan", "jeneng raja", "wilangan siji", "sandhangan pangkon mung", "A", "Bebasan ngiaske kahanan.", "bebasan"],
  ["Saloka ngemu …", "pepindhan watak/pribadi kanthi kiasan", "resep masakan", "tatacara salat", "rumus matématika", "A", "Saloka kiasan watak utawa pribawa.", "saloka"],
  ["Tembang macapat tuladhane …", "Dhandhanggula, Kinanthi, Pucung", "soneta Inggris", "pantun Melayu wajib", "haiku Jepang", "A", "Macapat kalebu Dhandhanggula lan liya-liyane.", "tembang"],
  ["Aksara Jawa uga diarani …", "hanacaraka / carakan", "rongorongo", "hieroglif", "hangul", "A", "Carakan hanacaraka minangka aksara Jawa.", "aksara"],
  ["Ha-na-ca-ra-ka ngemu crita …", "Aji Saka lan abdi", "Ramayana India murni tanpa lokal", "Perang Dunia", "Olimpiade", "A", "Legenda Aji Saka kagandheng hanacaraka.", "aksara"],
  ["Sandhangan pepet nulis swara …", "é", "e pepet (e)", "o", "i", "B", "Pepet kanggo swara e pepet.", "aksara"],
  ["Unggah-ungguh yen guneman karo guru kudu …", "ngoko kasar", "basa krama kang sopan", "bilingual Inggris wajib", "meneng wae tanpa basa", "B", "Guru dihormati nganggo krama.", "unggah-ungguh"],
  ["Tembung 'kula' tegese …", "kowe", "aku (krama)", "dheweke", "kita kabeh kasar", "B", "Kula = aku ing krama.", "kosakata"],
  ["Tembung 'panjenengan' tegese …", "aku", "kowe/panjenengan (krama inggil)", "dheweke kasar", "anak", "B", "Panjenengan ngurmati lawan guneman.", "kosakata"],
  ["Wayang kulit kalebu …", "budaya Jawa/seni pedhalangan", "olahraga modern", "basa Latin", "ilmu fisika", "A", "Wayang minangka warisan budaya.", "budaya"],
  ["Gamelan iku …", "prangkat musik Jawa/Bali", "jinis tarian Saman", "aksara Pallawa", "panganan", "A", "Gamelan: ensambel musik.", "budaya"],
  ["Batik minangka …", "kain kanthi ragam hias tutup celup", "alat musik", "tembang gedhe", "wilangan sengkala", "A", "Batik kalebu warisan budaya.", "budaya"],
  ["Keris kalebu …", "tosan aji/budaya materi Jawa", "alat nulis aksara", "jinis tembang", "sandhangan aksara", "A", "Keris minangka tosan aji.", "budaya"],
  ["Crita rakyat Jaka Tarub kalebu …", "legenda/folklor", "laporan ilmiah", "undang-undang", "kamus Inggris", "A", "Jaka Tarub kalebu folklor.", "cerita rakyat"],
  ["Roro Jonggrang kagandheng candhi …", "Borobudur mung", "Prambanan (legi)", "Mendut mung", "Sewu mung tanpa crita", "B", "Crita Roro Jonggrang kagandheng Prambanan.", "cerita rakyat"],
  ["Panji minangka …", "tokoh sastra/crita Jawa Timur-an", "nabi", "aksara rekan", "sandhangan layar", "A", "Panji misuwur ing sastra Jawa.", "sastra"],
  ["Serat Wedhatama kalebu karya …", "KGPAA Mangkunagara IV", "Shakespeare", "Homer", "Confucius", "A", "Wedhatama saka Mangkunagara IV.", "sastra"],
  ["Tembang Pocung nduweni guru gatra …", "4", "7", "10", "12", "A", "Pucung guru gatrane 4.", "tembang"],
  ["Guru wilangan ing tembang tegese …", "wilangan wanda saben gatra", "jeneng dalang", "jinis batik", "wilayah kraton mung", "A", "Guru wilangan: jumlah wanda.", "tembang"],
  ["Guru lagu tegese …", "tibaning swara pungkasan gatra", "alat gamelan", "jeneng panakawan", "aksara murda mung", "A", "Guru lagu: vokal pungkasan.", "tembang"],
  ["Tembung 'dhahar' kanggo …", "mangan (krama inggil kanggo liyan)", "turu kasar", "ngombe ngoko", "mlaku ngoko", "A", "Dhahar ngurmati liyan nalika mangan.", "kosakata"],
  ["Aja nggunakake krama inggil kanggo …", "awak dhewe", "bapak/ibu", "guru", "tamu sepuh", "A", "Krama inggil ora kanggo nyebut awak dhewe.", "unggah-ungguh"],
  ["Tembung 'kagungan' tegese …", "duwe (krama inggil)", "ilang", "tuku", "adol", "A", "Kagungan = gadhah/duwe kang ngormati.", "kosakata"],
  ["Panakawan tuladhane …", "Semar, Gareng, Petruk, Bagong", "Arjuna mung", "Rahwana mung", "Hanoman mung", "A", "Panakawan misuwur papat mau.", "budaya"],
  ["Bahasa Jawa ngoko lugu beda karo ngoko alus amarga …", "tingkat kesopanan/campuran krama", "aksara beda total", "ora bisa ditulis", "mung angka", "A", "Ngoko alus nyampur undha-usuk.", "unggah-ungguh"],
  ["Ukara 'Kula badhe tindak' tegese …", "Aku arep mangan", "Aku arep lunga", "Aku arep turu", "Aku arep ngombe", "B", "Tindak = lunga.", "kosakata"],
  ["Tembung 'nyuwun sewu' kanggo …", "nyuwun pangapunten/atur pangajak", "menehi dhahar", "ngina", "ngongkon kasar", "A", "Nyuwun sewu minangka basa hormat panyuwun.", "unggah-ungguh"],
  ["Candra sengkala iku …", "tinulis taun kanthi tetembungan kias", "jinis gending mung", "alat wayang", "sandhangan cakra", "A", "Sengkala: taun kiasan.", "sastra"],
  ["Geguritan iku …", "puisi Jawa modern/tradhisi anyar", "undang-undang", "resep jamu wajib", "peta", "A", "Geguritan minangka gegayutan puisi.", "sastra"],
  ["Purwakanthi tegese …", "pohaning swara/sastra ing ukara", "jinis keris", "wilangan guru gatra mung", "nami kraton", "A", "Purwakanthi: paesan swara.", "sastra"],
  ["Wangsalan iku …", "cak-cakan teka-teki kiasan", "alat gamelan", "aksara pasangan", "tembang tengahan mung", "A", "Wangsalan minangka cangkriman kias.", "sastra"],
  ["Tembung 'pados' krama saka …", "golek", "turu", "mangan", "ngombe", "A", "Golek → pados.", "kosakata"],
  ["Tembung 'ngendika' kanggo …", "ngomong (krama inggil kanggo liyan)", "mlaku ngoko", "turu ngoko", "nulis Latin", "A", "Ngendika ngurmati wong kang guneman.", "kosakata"],
  ["Aksara pasangan digunakake kanggo …", "mateni konsonan/nggawa konsonan sabanjure", "nulis angka Arab mung", "ngganti tembang", "ngukur irama", "A", "Pasangan mateni aksara sakdurunge.", "aksara"],
  ["Pada luhur utawa pada lungsi gegayutan …", "tandha wacan ing naskah Jawa", "jinis batik parang mung", "nama panakawan", "wilangan macapat", "A", "Pada minangka tandha wacan.", "aksara"],
  ["Tembung 'kula nuwun' kanggo …", "kethuk lawang/atur tabuh kang sopan", "ngina", "ngongkon mangan", "ngatur judhul buku", "A", "Kula nuwun basa kethuk lawang.", "unggah-ungguh"],
  ["Srimpi lan bedhaya kalebu …", "beksan kraton", "jinis pari", "aksara murda", "tembung entar mung", "A", "Keduanya tarian klasik.", "budaya"],
  ["Tembung entar tegese …", "tembung kiasan/idiomatik Jawa", "aksara swara", "wilangan siji", "nama gunung", "A", "Tembung entar: makna kias.", "kosakata"],
  ["Pepindhan migunakake tembung …", "kaya, lir, kadya, lsp.", "mung angka", "mung Latin", "mung Inggris", "A", "Pepindhan: perbandingan.", "sastra"],
  ["Crita Kancil kalebu …", "fabel/crita kewan", "biografi nabi", "laporan kimia", "pidato Inggris", "A", "Kancil minangka fabel.", "cerita rakyat"],
  ["Upacara tedhak siten kalebu …", "adat Jawa tumrap bayi", "salat Id", "olimpiade", "ujian nasional", "A", "Tedhak siten adat Jawa.", "budaya"],
  ["Slametan minangka …", "kenduri/adat syukur Jawa", "jinis aksara", "tembang megatruh mung", "alat wayang", "A", "Slametan: kenduri.", "budaya"],
  ["Tembung 'mangayubagya' tegese …", "nyuwun pangapunten", "ngucapake slamet/rahayu", "ngina", "ngongkon kasar", "B", "Mangayubagya: atur slamet.", "kosakata"],
  ["Basa kedhaton kanggo …", "lingkungan kraton", "pasar iwak mung", "sekolah Inggris mung", "stadion", "A", "Basa kedhaton kanggo palemahan kraton.", "unggah-ungguh"],
  ["Arjuna ing pewayangan kerep minangka …", "ksatria alus", "raksasa", "panakawan mung", "dewa langit Yunani", "A", "Arjuna ksatria Pandhawa.", "budaya"],
  ["Semar minangka …", "panakawan sepuh kang bijak", "prabu Kurawa", "raja raksasa", "naga naga", "A", "Semar pinunjul minangka panakawan.", "budaya"],
  ["Tembung 'nyuwun tulung' iku basa …", "panyuwun kang sopan", "perintah kasar", "ngoko kasar banget", "basa Walanda", "A", "Nyuwun tulung: panyuwun hormat.", "unggah-ungguh"],
  ["Aksara murda kanggo …", "tulisan jeneng/gelar kang dihormati", "mateni kabeh konsonan", "nulis tembang mung", "ngukur irama kendhang", "A", "Murda kanggo honorer jeneng.", "aksara"],
  ["Tembang Maskumambang watake …", "nelangsa/keluh kesah", "bungah ria murni tanpa sedhih", "perang gending mung", "olahraga", "A", "Maskumambang asring watak nelangsa.", "tembang"],
  ["Tembang Kinanthi kerep kanggo …", "pitutur/tresna kang alus", "prentahe perang kasar mung", "ngitung pajak", "nulis rumus", "A", "Kinanthi kerep pitutur.", "tembang"],
  ["Candi Borobudur dumunung ing …", "Jawa Tengah (Magelang)", "Bali", "Aceh", "Papua", "A", "Borobudur ing Magelang, Jateng.", "budaya"],
  ["Prinsip gotong royong ing budaya Jawa tegese …", "guyub gotong kerja bebarengan", "mung menang dhewe", "ngina tangga", "mutung silaturahmi", "A", "Gotong royong: bebarengan.", "budaya"],
  ["Tembung 'sugeng enjing' tegese …", "sugeng dalu", "sugeng enjing/selamat pagi", "matur nuwun", "nyuwun pangapunten", "B", "Sugeng enjing = selamat pagi.", "kosakata"],
  ["Tembung 'matur nuwun' tegese …", "nyuwun ngapura", "hatur thank you/ matur nuwun", "ayoh mangan", "lunga saiki", "B", "Matur nuwun = terima kasih.", "kosakata"],
  ["Tembung 'pangapunten' tegese …", "pamitan", "nyuwun ngapura", "panganan", "tembang", "B", "Pangapunten = maaf.", "kosakata"],
  ["Unen-unen 'alon-alon waton kelakon' tegese …", "cepatsalah wajib", "mbaka sithik nanging kelakon", "ora usah nyambut gawe", "nggege mongso kudu", "B", "Alon nanging tumekan tujuan.", "paribasan"],
  ["Unen-unen 'jer basuki mawa bea' tegese …", "sukses mbutuhake bea/pengorbanan", "gratis kabeh", "ora perlu usaha", "mung nasib", "A", "Kasuksesan mawa ragad/usaha.", "paribasan"],
  ["Tembung 'pados dhuwit' tegese …", "golek dhuwit", "tibo dhuwit", "ngobong dhuwit", "ngutang wajib", "A", "Pados = golek.", "kosakata"],
  ["Sastra piwulang kerep isine …", "pitutur luhur", "resep kimia", "kode program", "peta cuaca mung", "A", "Sastra piwulang: piwulang susila.", "sastra"],
  ["Tembung 'kagungan putra' luwih sopan tinimbang …", "nduwe anak (ngoko langsung marang sepuh)", "panjenengan", "kula", "nderek", "A", "Kagungan putra krama inggil kanggo liyan.", "unggah-ungguh"],
  ["Nalika nyapa sepuh, luwih trep …", "kowe tuku apa?", "panjenengan mundhut punapa?", "hei, kowe!", "eh botah", "B", "Mundhut/panjenengan luwih krama.", "unggah-ungguh"],
  ["Tembung 'mundhut' krama saka …", "tuku/njupuk miturut konteks", "turu", "ngombe kasar", "mlaku ngoko", "A", "Mundhut asring kanggo njupuk/tuku kang ngormati.", "kosakata"],
  ["Gending iku gegayutan …", "musik/karawitan", "olah raga bal-balan mung", "aksara rekan", "pajak", "A", "Gending: lagu karawitan.", "budaya"],
  ["Tembung 'nderek' tegese …", "melu/ndherek", "nampik", "ngina", "ngongkon kasar", "A", "Nderek = melu kanthi sopan.", "kosakata"],
  ["Aksara swara tuladhane …", "a, i, u, é, o (swara mandiri)", "pasangan ha", "layar mung", "cecak mung", "A", "Aksara swara nulis vokal mandiri.", "aksara"],
  ["Pitutur luhur Jawa nengenake …", "andhap asor lan tepo sliro", "sombong", "ngina liyan", "mutung guyub", "A", "Andhap asor lan tepa slira diajarke.", "budaya"],
];

function jawaBank() {
  return JAWA.slice(0, 75).map((row, i) =>
    item(`jawa-${String(i + 1).padStart(3, "0")}`, "BAHASA_JAWA", row[7], i, row[0], row[1], row[2], row[3], row[4], row[5], row[6]),
  );
}

const pai = paiBank();
const math = mathBank();
const eng = engBank();
const jawa = jawaBank();
const all = [...pai, ...math, ...eng, ...jawa];

if (pai.length !== 75 || math.length !== 75 || eng.length !== 75 || jawa.length !== 75 || all.length !== 300) {
  throw new Error(`counts pai=${pai.length} math=${math.length} eng=${eng.length} jawa=${jawa.length} total=${all.length}`);
}

const ids = new Set(all.map((q) => q.id));
const texts = new Set(all.map((q) => q.question));
if (ids.size !== 300 || texts.size !== 300) {
  throw new Error(`dup id=${ids.size} text=${texts.size}`);
}
for (const q of all) {
  if (!["A", "B", "C", "D"].includes(q.correctAnswer) || !q.explanation || !q.optionA || !q.optionB || !q.optionC || !q.optionD) {
    throw new Error(`invalid ${q.id}`);
  }
}

const dir = dirname(fileURLToPath(import.meta.url));
const dest = join(dir, "../mmo/data/sma-edu-questions.json");
writeFileSync(dest, JSON.stringify(all, null, 0));
console.log("wrote", dest, all.length);
