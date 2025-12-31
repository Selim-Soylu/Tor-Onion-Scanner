# Tor Onion Scanner

Bu proje, Tor ağı üzerinden .onion adreslerini tarayan Go tabanlı bir uygulamadır.
Uygulama SOCKS5 proxy (127.0.0.1:9050) kullanarak hedeflere istek atar, erişilebilen
sayfaların HTML içeriğini kaydeder ve chromedp ile ekran görüntüsü alır.

----------------------------------------------------------------

GEREKSİNİMLER

- Go 1.20 veya üzeri
- Tor servisi
- Chromium veya Google Chrome

----------------------------------------------------------------

ÇALIŞTIRMA

1) Tor servisini başlat:
tor

2) Proje dizininde bağımlılıkları indir:
go mod tidy

3) Programı çalıştır:
go run scanner.go targets.yaml

----------------------------------------------------------------

TARAMA SÜRECİ

Program çalışırken terminalde anlık loglar üretilir.

Örnek çıktılar:

[INFO] Scanning: http://example.onion -> SUCCESS  
[ERR] Scanning: http://badurl.onion -> TIMEOUT  

----------------------------------------------------------------

TOR IP DOĞRULAMA

Program, isteklerin Tor ağı üzerinden gönderildiğini doğrulamak için
check.torproject.org adresini kullanır.

Başarılı durumda terminal çıktısı:

[INFO] Tor connection verified

----------------------------------------------------------------

ÇIKTILAR

Tarama tamamlandıktan sonra aşağıdaki klasörler otomatik oluşur:

- output/html/
- output/screenshots/
