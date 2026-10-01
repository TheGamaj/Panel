<p align="center">
 <a href="../README.md">English</a> /
 <a href="./README-fa.md">فارسی</a> /
 <a href="./README-ru.md">Русский</a> /
 <a href="./README-zh-cn.md">简体中文</a>
</p>

# Панель Gamaj

**Gamaj (گمج)** — это self-hosted платформа управления прокси-аккаунтами на
базе Xray: веб-панель, полный REST API, CLI и распределение по нескольким
нодам. Всё это — один нативный бинарник и служба systemd.

Текущая версия: **`is.0.0.1`** · Порт по умолчанию: **616**

<p align="center">
  <a href="https://t.me/TheGamaj">Канал в Telegram</a> ·
  <a href="https://t.me/GamajGP">Группа поддержки</a> ·
  <a href="https://t.me/AsliCode">Канал с кодом</a>
</p>

## Возможности

- **Веб-панель** (React + Chakra UI), светлая и тёмная тема, четыре языка: английский, فارسی, русский, 简体中文
- **REST API** для всех действий панели и **API-ключи** для ботов и интеграций
- **Несколько нод** с управлением по gRPC и взаимной проверкой сертификатов
- Протоколы: **VLESS**, **VMess**, **Trojan**, **Shadowsocks**, а на нодах также OpenVPN, WireGuard, L2TP, PPTP, SSTP и Remote Access
- **Несколько пользователей на один inbound**, несколько inbound на пользователя, fallback на одном порту
- **Ограничения трафика, срока действия и одновременных подключений**, периодический сброс трафика (день/неделя/…)
- **Ссылки подписки** для v2ray-клиентов, Clash и sing-box, автоматические ссылки и QR-коды
- **Интеграция с Telegram**: отчёты, резервные копии и уведомления пользователям
- **Резервное копирование и восстановление** (база и полный архив) с отправкой в Telegram по расписанию
- **Мониторинг**: CPU, память, диск, онлайн-пользователи, трафик каждой ноды, логи Xray
- **Несколько администраторов** с ролями `full_access`, `sudo`, `standard`, `seller`, 2FA, персональными лимитами и API-ключами
- **Хостинг внешних приложений**: установка и управление сторонними панелями на том же сервере
- **HAProxy и управление хостами**: шаблоны, адреса, сертификаты и действия на нодах
- **Полностью собственный бренд**: одно имя, один порт, без контейнерного слоя

## Требования

- Linux-сервер с systemd (Debian/Ubuntu, CentOS/Rocky/Alma, Fedora, Arch, Alpine) и права root
- Архитектуры: `linux-386`, `linux-amd64`, `linux-arm64`, `linux-armv5`, `linux-armv6`, `linux-armv7`, `linux-s390x`
- База данных: SQLite (по умолчанию), MySQL или MariaDB

## Установка

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

Dev-канал или конкретный релиз:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --dev
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --version is.0.0.1
```

Выбор базы данных:

```bash
... | sudo bash -s -- install --database sqlite
... | sudo bash -s -- install --database mysql
... | sudo bash -s -- install --database mariadb
```

Не запускайте установщик как `sudo bash -c "$(curl ...)"`: тело скрипта может
превысить лимит длины аргумента в Linux. Всегда передавайте вывод curl в bash.

Gamaj устанавливается как нативный бинарник со службой systemd. Контейнерной
установки и переключателя режимов нет: один путь установки и все возможности.

## После установки

| Что | Где |
|---|---|
| Файлы панели | `/opt/gamaj` |
| Конфигурация | `/opt/gamaj/.env` |
| Данные и база | `/var/lib/gamaj` |
| Служба | `gamaj.service` |

Создайте первого администратора:

```bash
sudo gamaj cli admin create --role full_access
```

Открыть панель:

- С доменом и TLS: `https://YOUR_DOMAIN:616/dashboard/`
- Для локального теста через SSH-туннель: `ssh -L 616:localhost:616 user@serverip`, затем `http://localhost:616/dashboard/`

Полная пошаговая инструкция, включая TLS и firewall, — в
[INSTALL.md](INSTALL.md).

## Управление

```bash
sudo gamaj status          # состояние службы
sudo gamaj restart         # перезапуск
sudo gamaj logs            # логи
sudo gamaj backup          # резервная копия
sudo gamaj update          # обновление
sudo gamaj update --dev    # обновление до dev-канала
sudo gamaj core-update     # обновление Xray-core
sudo gamaj edit-env        # редактирование /opt/gamaj/.env
sudo gamaj ssl             # управление сертификатами
sudo gamaj uninstall       # удаление
```

## Конфигурация

Все настройки находятся в `/opt/gamaj/.env`, все ключи начинаются с `GAMAJ_`.
Полный список — в [`.env.example`](../.env.example). Основные:

| Ключ | Значение | По умолчанию |
|---|---|---|
| `GAMAJ_DATABASE_URL` | Подключение к базе (обязательно) | — |
| `GAMAJ_HOST` | Адрес прослушивания | `0.0.0.0` |
| `GAMAJ_PORT` | Порт панели | `616` |
| `GAMAJ_DATA_DIR` | Каталог данных | `/var/lib/gamaj` |
| `GAMAJ_XRAY_JSON` | Файл конфигурации Xray | `/var/lib/gamaj/xray_config.json` |
| `GAMAJ_USER_AUTODELETE_DAYS` | Удалять истёкших пользователей через N дней | `0` (выключено) |
| `GAMAJ_JWT_ACCESS_TOKEN_EXPIRE_MINUTES` | Время жизни сессии панели | `1440` |
| `GAMAJ_WEBHOOK_ADDRESS` | Адреса webhook-уведомлений | — |

## Добавление ноды

На другом сервере:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

Затем в панели откройте **Nodes → добавить ноду**. Панель выдаёт сертификат и
конфигурацию, а нода сама отправляет трафик, онлайн-пользователей и выполняет
действия на хосте.

## Установка бота

Telegram-бот живёт в отдельном репозитории: [TheGamaj/Bot](https://github.com/TheGamaj/Bot) — работает только через API панели. У самого бота платёжного шлюза нет: шлюз принадлежит панели, бот использует его через API.

Рекомендуемый путь: откройте **Applications** в панели и установите **Gamaj Bot** — панель сама скачает бинарник, создаст ключ `gm_...`, запишет конфиг и зарегистрирует webhook.

Ручная установка:

```bash
git clone https://github.com/TheGamaj/Bot.git
cd Bot
go build -o gamajbot ./cmd/gamajbot
sudo ./scripts/install.sh
sudo nano /opt/gamaj-bot/.gamaj-bot.json   # panel_url, api_key, bot_token, admin_id
sudo systemctl restart gamaj-bot
```

## Командная строка

`gamaj cli` управляет пользователями, администраторами, inbound и базой данных:
например `gamaj cli user list`, `gamaj cli admin set-password`,
`gamaj cli migrate up`. Справочник: [cli/README.md](cli/README.md).

## Разработка

```bash
cd dashboard && npm ci && VITE_BASE_API=/api/ npm run build && cd ..
bash scripts/build_binary.sh        # dist/gamaj-server и dist/gamaj-cli

go build ./... && go vet ./internal/... ./cmd/...
go test ./internal/app/system/ ./internal/app/api/ ./internal/gateway/ \
        ./internal/app/backup/ ./internal/app/migrations/ ./cmd/... \
        ./internal/platform/envguard/

sudo bash scripts/tests/e2e-binary-install.sh   # сквозная проверка установки
```

Тест `go test ./internal/platform/envguard/` ломает сборку, если в дереве
появится переменная окружения без префикса `GAMAJ_` или устаревшее имя.

## Документация

- [INSTALL.md](INSTALL.md) — установка и обновление
- [cli/README.md](cli/README.md) — справочник CLI
- [dashboard/README.md](dashboard/README.md) — заметки о панели
- [scripts/gamaj/README.md](scripts/gamaj/README.md) — скрипты установки
- [CONTRIBUTING.md](CONTRIBUTING.md) — правила участия
- [tutorials/](../tutorials/) — пошаговые руководства внутри панели

## Лицензия

GNU Affero General Public License v3.0 — файл [LICENSE](../LICENSE).
