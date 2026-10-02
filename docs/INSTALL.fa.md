# راهنمای نصب و به‌روزرسانی پنل گمج

نسخه: **`is.0.0.1`** · پورت پیش‌فرض: **۶۱۶**

گمج پنل به‌صورت باینری بومی با سرویس systemd نصب می‌شود. نصب کانتینری وجود
ندارد: دستور `install` هیچ‌وقت ایمیج نمی‌گیرد، به docker نیاز ندارد و نمی‌پرسد
«چه حالت نصبی؟».

<p align="center">
  <a href="https://t.me/TheGamaj">کانال</a> ·
  <a href="https://t.me/GamajGP">گروه پشتیبانی</a> ·
  <a href="https://t.me/AsliCode">کانال کد</a>
</p>

## ۱. پیش‌نیازها

| مورد | مقدار |
|---|---|
| سیستم‌عامل | لینوکس با systemd: دبیان/اوبونتو، سنت‌اواس/راکی/آلما، فدورا، آرچ، آلپاین |
| دسترسی | root یا کاربر با sudo |
| پورت | **۶۱۶** برای داشبورد و API |
| دیتابیس | SQLite (پیش‌فرض)، MySQL یا MariaDB |
| معماری‌ها | `linux-386`, `linux-amd64`, `linux-arm64`, `linux-armv5`, `linux-armv6`, `linux-armv7`, `linux-s390x` |
| برای ساخت از سورس | Go 1.25+ و Node.js 20+ با npm |

## ۲. نصب با اسکریپت رسمی (پیشنهادی)

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

گزینه‌ها:

```bash
... | sudo bash -s -- install --dev                  # کانال dev
... | sudo bash -s -- install --version is.0.0.1     # نسخه‌ی مشخص
... | sudo bash -s -- install --database sqlite      # یا mysql / mariadb
```

نکته‌ها:

- هرگز `sudo bash -c "$(curl ...)"` استفاده نکنید؛ بدنه‌ی اسکریپت می‌تواند از
  محدودیت طول آرگومان لینوکس بزرگ‌تر شود. همیشه خروجی را به bash بدهید.
- وقتی stdin ترمینال نباشد (اتوماسیون، CI، `curl | bash`) نصب‌کننده همه‌ی
  انتخاب‌ها را خودکار انجام می‌دهد و حالت نصب باینری است.
- اگر ریلیز برای معماری سرور دارایی (asset) نداشته باشد، نصب با پیام روشن
  (نام دارایی موردانتظار، دارایی‌های موجود و معماری‌های پشتیبانی‌شده) متوقف
  می‌شود و نصب نیمه‌کاره باقی نمی‌گذارد.

## ۳. نصب دستی از سورس

```bash
git clone https://github.com/TheGamaj/Panel.git Gamaj-Panel
cd Gamaj-Panel/dashboard && npm ci && VITE_BASE_API=/api/ npm run build
cp build/index.html build/404.html
cd ..

bash scripts/build_binary.sh      # dist/gamaj-server و dist/gamaj-cli

sudo mkdir -p /opt/gamaj/bin /var/lib/gamaj
sudo install -m 0755 dist/gamaj-server /opt/gamaj/bin/gamaj-server
sudo install -m 0755 dist/gamaj-cli /opt/gamaj/bin/gamaj-cli
sudo cp .env.example /opt/gamaj/.env
sudo nano /opt/gamaj/.env         # حداقل GAMAJ_DATABASE_URL را تنظیم کنید
```

فایل `/etc/systemd/system/gamaj.service`:

```ini
[Unit]
Description=Gamaj Panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/gamaj
Environment=GAMAJ_APP_DIR=/opt/gamaj
Environment=GAMAJ_ENV_FILE=/opt/gamaj/.env
Environment=GAMAJ_INSTALL_MODE=binary
Environment=GAMAJ_DATA_DIR=/var/lib/gamaj
ExecStart=/opt/gamaj/bin/gamaj-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload && sudo systemctl enable --now gamaj
```

## ۴. نصب آفلاین

باینری‌ها را روی ماشین متصل بسازید (بخش ۳) و سپس روی سرور مقصد نصب‌کننده را به
فایل‌های محلی وصل کنید:

```bash
sudo GAMAJ_BINARY_SERVER_OVERRIDE="$PWD/gamaj-server" \
     GAMAJ_BINARY_CLI_OVERRIDE="$PWD/gamaj-cli" \
     GAMAJ_BINARY_OVERRIDE_VERSION=is.0.0.1 \
     bash gamaj.sh install --database sqlite
```

## ۵. اولین اجرا

```bash
sudo gamaj cli admin create --role full_access   # ادمین full_access
sudo gamaj status
sudo gamaj logs
```

داشبورد:

- با دامنه و SSL: `https://YOUR_DOMAIN:616/dashboard/`
- با تونل SSH برای تست لوکال: `ssh -L 616:localhost:616 user@serverip` و سپس
  `http://localhost:616/dashboard/`

### گواهی SSL

```bash
sudo gamaj ssl issue --email you@example.com --domains panel.example.com
sudo gamaj ssl issue --email you@example.com --ip-address 203.0.113.10
sudo gamaj ssl issue --email you@example.com --domains 203.0.113.10 --provider self-signed
sudo gamaj ssl renew
```

یا مسیر گواهی موجود خودتان را در `.env` بگذارید:

```dotenv
GAMAJ_SSL_CERTFILE=/etc/letsencrypt/live/panel.example.com/fullchain.pem
GAMAJ_SSL_KEYFILE=/etc/letsencrypt/live/panel.example.com/privkey.pem
```

### فایروال

فقط پورت پنل باید از بیرون باز باشد:

```bash
sudo ufw allow 616/tcp
# یا
sudo firewall-cmd --permanent --add-port=616/tcp && sudo firewall-cmd --reload
```

## ۶. افزودن نود

روی سرور نود:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

سپس در پنل، بخش **Nodes** را باز کنید و آدرس و پورت کنترل نود (پیش‌فرض
`62050`) را ثبت کنید. پنل گواهی و کانفیگ را می‌سازد و نود مصرف، کاربران آنلاین
و عملیات هاست را گزارش می‌کند.

## ۷. به‌روزرسانی و حذف

```bash
sudo gamaj update                 # آخرین نسخه‌ی منتشرشده
sudo gamaj update --dev           # کانال dev
sudo gamaj update --version is.0.0.1
sudo gamaj uninstall
```

به‌روزرسانی باینری‌ها را جایگزین می‌کند، `/opt/gamaj/.env` و `/var/lib/gamaj`
را دست نمی‌زند، مایگریشن‌ها را در استارت اجرا می‌کند و سرویس را ری‌استارت
می‌کند.

## ۸. عیب‌یابی

| نشانه | بررسی |
|---|---|
| پنل روی ۶۱۶ جواب نمی‌دهد | `systemctl status gamaj`، `journalctl -u gamaj -n 100`، `ss -ltn \| grep 616` |
| نصب‌کننده می‌گوید دارایی پیدا نشد | ریلیز باید برای معماری شما دارایی داشته باشد؛ نصب با پیام دقیق متوقف می‌شود |
| حلقه‌ی ورود در داشبورد | `GAMAJ_DATABASE_URL` و دسترسی نوشتن سرویس به `GAMAJ_DATA_DIR` |
| خطای گواهی | `sudo gamaj ssl renew` و مسیرهای `GAMAJ_SSL_CERTFILE`/`GAMAJ_SSL_KEYFILE` |
| نود در پنل آفلاین است | باز بودن پورت کنترل بین دو سرور و تطابق گواهی نود با رکورد پنل |

## ۹. تست سرتاسری نصب

```bash
sudo bash scripts/tests/e2e-binary-install.sh
```

این اسکریپت داشبورد و باینری‌ها را می‌سازد، در یک sandbox نصب می‌کند، سرویس را
بالا می‌آورد و اگر پنل روی پورت ۶۱۶ سالم نباشد یا لاگین ادمین از API کار نکند،
شکست می‌دهد. روی ویندوز/مک با پیام `SKIP` رد می‌شود.
