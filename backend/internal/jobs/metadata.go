package jobs

import (
	"encoding/json"
	"fmt"
	"strings"
)

var Categories = []Option{{"engineering", "IT & Engineering"}, {"office", "Administrasi & Korporat"}, {"hospitality", "Hotel & Pariwisata"}, {"nursing-care", "Kaigo"}, {"building-cleaning", "Pembersihan Gedung"}, {"manufacturing", "Manufaktur"}, {"construction", "Konstruksi"}, {"agriculture", "Pertanian & Peternakan"}, {"fishery", "Perikanan & Budi Daya"}, {"food-service", "Restoran"}, {"food-manufacturing", "Produksi Makanan"}, {"logistics", "Logistik & Gudang"}, {"transportation", "Transportasi"}, {"education", "Pendidikan"}, {"creative", "Desain & Kreatif"}, {"sales", "Sales & Marketing"}, {"finance", "Keuangan"}, {"healthcare", "Kesehatan"}, {"general", "Lainnya"}}
var EmploymentTypes = []Option{{"full-time", "Penuh waktu"}, {"part-time", "Paruh waktu"}, {"contract", "Kontrak"}, {"temporary", "Sementara"}, {"internship", "Magang"}, {"unknown", "Tidak dicantumkan"}}
var JapaneseLevels = []Option{{"N1", "JLPT N1"}, {"N2", "JLPT N2"}, {"N3", "JLPT N3"}, {"N4", "JLPT N4"}, {"N5", "JLPT N5"}, {"not-required", "Jepang tidak diwajibkan"}, {"unknown", "Tidak dicantumkan"}}
var VisaOptions = []Option{{"sponsored", "Sponsorship disebutkan"}, {"ssw", "SSW disebutkan"}, {"work-permit-required", "Izin kerja wajib dimiliki"}, {"unknown", "Tidak dicantumkan"}}
var Sectors = []Option{{"nursing-care", "Kaigo"}, {"building-cleaning", "Pembersihan Gedung"}, {"linen-supply", "Pengolahan Linen"}, {"industrial-manufacturing", "Manufaktur Produk Industri"}, {"construction", "Konstruksi"}, {"shipbuilding", "Kapal & Mesin Kapal"}, {"automobile-maintenance", "Perawatan Mobil"}, {"aviation", "Penerbangan"}, {"accommodation", "Perhotelan"}, {"road-transport", "Transportasi Kendaraan"}, {"railway", "Perkeretaapian"}, {"logistics-warehouse", "Gudang Logistik"}, {"agriculture", "Pertanian"}, {"fishery", "Perikanan"}, {"food-manufacturing", "Produksi Makanan"}, {"food-service", "Layanan Makanan"}, {"forestry", "Kehutanan"}, {"wood-industry", "Industri Kayu"}, {"resource-recycling", "Daur Ulang"}}
var Prefectures = func() []Option {
	names := strings.Split("Hokkaido,Aomori,Iwate,Miyagi,Akita,Yamagata,Fukushima,Ibaraki,Tochigi,Gunma,Saitama,Chiba,Tokyo,Kanagawa,Niigata,Toyama,Ishikawa,Fukui,Yamanashi,Nagano,Gifu,Shizuoka,Aichi,Mie,Shiga,Kyoto,Osaka,Hyogo,Nara,Wakayama,Tottori,Shimane,Okayama,Hiroshima,Yamaguchi,Tokushima,Kagawa,Ehime,Kochi,Fukuoka,Saga,Nagasaki,Kumamoto,Oita,Miyazaki,Kagoshima,Okinawa", ",")
	out := make([]Option, len(names))
	for i, name := range names {
		out[i] = Option{fmt.Sprintf("%02d", i+1), name}
	}
	return out
}()

func Metadata() map[string]any {
	return map[string]any{"categories": Categories, "prefectures": Prefectures, "employmentTypes": EmploymentTypes, "japaneseLevels": JapaneseLevels, "visaOptions": VisaOptions, "sectors": Sectors, "remoteOptions": []Option{{"onsite", "Di lokasi"}, {"hybrid", "Hybrid"}, {"remote", "Remote"}, {"unknown", "Tidak dicantumkan"}}}
}
func allowed(value string, options []Option) bool {
	for _, o := range options {
		if o.ID == value {
			return true
		}
	}
	return false
}
func CanonicalCategory(s string) string {
	switch s {
	case "it":
		s = "engineering"
	case "administration":
		s = "office"
	case "other":
		s = "general"
	}
	if allowed(s, Categories) {
		return s
	}
	return "general"
}

// The static source directory is a discovery resource, never an instruction to crawl.
func DecodeDirectory(raw []byte) ([]Source, error) {
	var list []Source
	err := json.Unmarshal(raw, &list)
	return list, err
}
