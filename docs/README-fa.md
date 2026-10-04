<p align="center">
 <a href="../README.md">English</a> /
 <a href="./README-fa.md">فارسی</a> /
 <a href="./README-ru.md">Русский</a> /
 <a href="./README-zh-cn.md">简体中文</a>
</p>

<p align="center">
  <img src="./assets/gamaj-logo.svg" alt="Gamaj" width="104" height="104">
</p>

<h1>GAMAJ</h1>

### Panel

**گمج** یک پلتفرم self-hosted برای مدیریت اکانت‌های پروکسی مبتنی بر Xray است:
داشبورد وب، API کامل REST، ابزار خط فرمان و پشتیبانی از چند نود — همه روی یک
باینری بومی با سرویس systemd.

نسخه‌ی فعلی: **`is.0.0.1`** · پورت پیش‌فرض: **۶۱۶**

<p align="center">
  <a href="https://t.me/TheGamaj">کانال تلگرام</a> ·
  <a href="https://t.me/GamajGP">گروه پشتیبانی</a> ·
  <a href="https://t.me/AsliCode">کانال کد</a>
</p>

## امکانات

- **داشبورد وب** (React + Chakra UI) با تم روشن/تیره و چهار زبان: انگلیسی، فارسی، روسی، چینی
- **API کامل REST** برای همه‌ی کارهای داشبورد + **کلید API** برای ربات‌ها و سرویس‌های جانبی
- **چند نود** با کنترل gRPC و گواهی دوطرفه و ثبت خودکار گواهی نود
- پروتکل‌ها: **VLESS**، **VMess**، **Trojan**، **Shadowsocks** و روی نودها: OpenVPN، WireGuard، L2TP، PPTP، SSTP و Remote Access
- **چند کاربر روی یک inbound**، چند inbound برای یک کاربر، fallback روی یک پورت
- **محدودیت حجم، تاریخ انقضا و تعداد اتصال همزمان**، با ریست دوره‌ای حجم (روزانه/هفتگی/…)
- **لینک اشتراک** برای کلاینت‌های v2ray، Clash و sing-box، همراه با لینک اشتراک و QR خودکار
- **اتصال به تلگرام**: گزارش‌ها، بکاپ خودکار و اطلاع‌رسانی به کاربران
- **بکاپ و بازیابی** (دیتابیس و آرشیو کامل) با امکان ارسال زمان‌بندی‌شده به تلگرام
- **مانیتورینگ**: CPU، رم، دیسک، کاربران آنلاین، مصرف هر نود و لاگ Xray
- **چند ادمین** با نقش‌های `full_access`، `sudo`، `standard` و `seller`، ورود دوعاملی (2FA)، محدودیت‌های اختصاصی و کلید API
- **میزبانی اپ‌های جانبی**: نصب و مدیریت اپ‌های ثالث روی همان سرور
- **HAProxy و مدیریت هاست**: قالب هاست، آدرس‌ها، گواهی‌ها و عملیات سطح هاست روی نودها
- **کاملاً برندشده و خودکفا**: یک نام، یک پورت، بدون هیچ لایه‌ی کانتینری

## پیش‌نیازها

- سرور لینوکس با systemd (دبیان/اوبونتو، سنت‌اواس/راکی/آلما، فدورا، آرچ، آلپاین) و دسترسی root
- معماری‌ها: `linux-386`, `linux-amd64`, `linux-arm64`, `linux-armv5`, `linux-armv6`, `linux-armv7`, `linux-s390x`
- دیتابیس: SQLite (پیش‌فرض)، MySQL یا MariaDB

## نصب

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

نسخه‌ی مشخص:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --version is.0.0.1
```

انتخاب دیتابیس:

```bash
... | sudo bash -s -- install --database sqlite
... | sudo bash -s -- install --database mysql
... | sudo bash -s -- install --database mariadb
```

نصب‌کننده را به‌شکل `sudo bash -c "$(curl ...)"` اجرا نکنید؛ بدنه‌ی اسکریپت
می‌تواند از محدودیت طول آرگومان لینوکس بزرگ‌تر باشد. همیشه خروجی curl را به
bash بدهید.

گمج فقط به‌صورت باینری بومی با سرویس systemd نصب می‌شود؛ نه Docker، نه حالت
نصب جایگزین. یک مسیر نصب، همه‌ی قابلیت‌ها.

## بعد از نصب

| مورد | مسیر |
|---|---|
| فایل‌های پنل | `/opt/gamaj` |
| تنظیمات | `/opt/gamaj/.env` |
| داده‌ها و دیتابیس | `/var/lib/gamaj` |
| سرویس | `gamaj.service` |

ساخت ادمین اصلی:

```bash
sudo gamaj cli admin create --role full_access
```

ورود به داشبورد:

- با دامنه و SSL: `https://YOUR_DOMAIN:616/dashboard/`
- برای تست لوکال با تونل SSH: `ssh -L 616:localhost:616 user@serverip` و سپس `http://localhost:616/dashboard/`

راهنمای گام‌به‌گام (شامل SSL و فایروال) در [INSTALL.md](INSTALL.md) است.

## مدیریت

```bash
sudo gamaj status          # وضعیت سرویس
sudo gamaj restart         # ری‌استارت
sudo gamaj logs            # لاگ زنده
sudo gamaj backup          # بکاپ
sudo gamaj update          # به‌روزرسانی
sudo gamaj core-update     # به‌روزرسانی Xray-core
sudo gamaj edit-env        # ویرایش /opt/gamaj/.env
sudo gamaj ssl             # مدیریت گواهی‌ها
sudo gamaj uninstall       # حذف گمج
```

## تنظیمات

همه‌ی تنظیمات در `/opt/gamaj/.env` هستند و همه‌ی کلیدها با `GAMAJ_` شروع
می‌شوند. فهرست کامل در [`.env.example`](../.env.example) است. پرکاربردترین‌ها:

| کلید | معنی | پیش‌فرض |
|---|---|---|
| `GAMAJ_DATABASE_URL` | اتصال دیتابیس (اجباری) | — |
| `GAMAJ_HOST` | آدرس bind گیت‌وی | `0.0.0.0` |
| `GAMAJ_PORT` | پورت گیت‌وی | `616` |
| `GAMAJ_DATA_DIR` | پوشه‌ی داده‌ها | `/var/lib/gamaj` |
| `GAMAJ_XRAY_JSON` | فایل کانفیگ Xray | `/var/lib/gamaj/xray_config.json` |
| `GAMAJ_USER_AUTODELETE_DAYS` | حذف کاربر منقضی بعد از چند روز | `0` (غیرفعال) |
| `GAMAJ_JWT_ACCESS_TOKEN_EXPIRE_MINUTES` | عمر نشست داشبورد | `1440` |
| `GAMAJ_WEBHOOK_ADDRESS` | آدرس وبهوک اطلاع‌رسانی | — |

## افزودن نود

روی سرور دیگر:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

بعد در داشبورد به **Nodes → افزودن نود** بروید. پنل گواهی و کانفیگ نود را
می‌سازد و نود خودش مصرف، کاربران آنلاین و عملیات هاست را گزارش می‌کند.

## نصب ربات

ربات تلگرام مخزن جداگانه‌ای دارد: [TheGamaj/Bot](https://github.com/TheGamaj/Bot) — فقط با API پنل کار می‌کند. تمام تنظیمات پرداخت در پنل نگهداری می‌شود و بات فقط از طریق API از آن‌ها استفاده می‌کند.

مسیر توصیه‌شده: در پنل به **Applications** بروید و **Gamaj Bot** را نصب کنید؛ پنل باینری را دانلود می‌کند، کلید API اختصاصی می‌سازد، کانفیگ بات را می‌نویسد و webhook را ثبت می‌کند.

نصب دستی:

```bash
git clone https://github.com/TheGamaj/Bot.git Gamaj-Bot
cd Gamaj-Bot
sudo ./scripts/install.sh    # گو را خودش نصب و باینری را خودش می‌سازد
sudo nano /opt/gamaj-bot/.gamaj-bot.json   # panel_url، api_key، bot_token، admin_id
sudo systemctl restart gamaj-bot
```

## خط فرمان

`gamaj cli` کاربران، ادمین‌ها، inbound‌ها و دیتابیس را از ترمینال مدیریت
می‌کند؛ مثلاً `gamaj cli user list`، `gamaj cli admin set-password`،
`gamaj cli migrate up`. مرجع کامل: [cli/README.md](cli/README.md).

## توسعه

```bash
cd dashboard && npm ci && VITE_BASE_API=/api/ npm run build && cd ..
bash scripts/build_binary.sh        # dist/gamaj-server و dist/gamaj-cli

go build ./... && go vet ./internal/... ./cmd/...
go test ./internal/app/system/ ./internal/app/api/ ./internal/gateway/ \
        ./internal/app/backup/ ./internal/app/migrations/ ./cmd/... \
        ./internal/platform/envguard/

sudo bash scripts/tests/e2e-binary-install.sh   # تست نصب سرتاسری روی پورت ۶۱۶
```

تست `go test ./internal/platform/envguard/` اگر جایی نام متغیر محیطی بدون
پیشوند `GAMAJ_` (یا نامی از دوران قبل از گمج) اضافه شود، بیلد را می‌شکند.

## مستندات

- [INSTALL.md](INSTALL.md) — راهنمای نصب و به‌روزرسانی
- [cli/README.md](cli/README.md) — مرجع خط فرمان
- [dashboard/README.md](dashboard/README.md) — نکات داشبورد
- [scripts/gamaj/README.md](scripts/gamaj/README.md) — اسکریپت‌های نصب
- [CONTRIBUTING.md](CONTRIBUTING.md) — قواعد مشارکت
- [tutorials/](../tutorials/) — آموزش‌های گام‌به‌گام داخل پنل

## لایسنس

GNU Affero General Public License v3.0 — فایل [LICENSE](../LICENSE).
