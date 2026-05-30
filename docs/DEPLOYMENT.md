# Развёртывание экосистемы Surface Kinetics

Пошаговая инструкция для стека **surface-atoms** (KMC на Go) + **Surface Kinetics Studio** (веб-редактор и API). Studio изменяет файлы симулятора на диске и вызывает `go build` / `go run`.

## Схема

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────┐
│  Браузер        │────▶│  Studio (Node)       │────▶│  surface-atoms  │
│  :5173 / :8085  │     │  API :3847 + React   │     │  Go KMC         │
└─────────────────┘     └──────────────────────┘     └─────────────────┘
                              │                              │
                              └──── SURFACE_ATOMS_PATH ──────┘
```

| Репозиторий | Назначение |
|-------------|------------|
| [surface-atoms](https://github.com/Kizerfifas/surface-atoms) | Симулятор, `config.yaml`, `configs/*.yaml`, папки `result …` |
| [surface-kinetics-studio](https://github.com/Kizerfifas/surface-kinetics-studio) | UI: схема, конфиг, запуск, результаты, sweep |

---

## 0. Требования

| Компонент | Версия | Назначение |
|-----------|--------|------------|
| **Go** | 1.23+ | Сборка и запуск симулятора |
| **Node.js** | 18+ | API и фронт Studio |
| **npm** | с Node | Зависимости Studio |
| **git** | любая | Клонирование |
| **nginx** | опционально | Доступ с другого ПК (Windows / Hyper-V) по порту **8085** |

Проверка:

```bash
go version    # go1.23.x или новее
node -v       # v18+
npm -v
```

---

## 1. Клонирование

Оба репозитория удобно положить **рядом** в одну папку:

```bash
mkdir -p ~/projects && cd ~/projects

git clone git@github.com:Kizerfifas/surface-atoms.git
git clone git@github.com:Kizerfifas/surface-kinetics-studio.git
```

Актуальная разработка — ветка **`dev`**:

```bash
cd ~/projects/surface-atoms && git checkout dev
cd ~/projects/surface-kinetics-studio && git checkout dev
```

Итоговая структура:

```
~/projects/
├── surface-atoms/           # симулятор
└── surface-kinetics-studio/ # веб-студия
```

---

## 2. Настройка surface-atoms

```bash
cd ~/projects/surface-atoms
go mod download
go build -o surface-atoms .
go test ./internal/scheme/... ./internal/simulator/...
```

Проверка **без Studio** (T = 300 K, время 1e-6 с):

```bash
go run . 300 1e-6
ls -d "result "* | tail -1
```

Должна появиться папка вида `result 2026-… T300K`.

Файлы по умолчанию:

- `config.yaml` — элементы, матрица, `schemePath`
- `configs/scheme_marinov.yaml` — пример схемы Marinov (в git)

Собственные схемы (`configs/scheme_my.yaml` и т.п.) остаются **локально** (см. `.gitignore`) и в коммиты не попадают.

---

## 3. Настройка Surface Kinetics Studio

```bash
cd ~/projects/surface-kinetics-studio
npm install
cd client && npm install && cd ..
```

Если `surface-atoms` **не** в `../surface-atoms`, задайте абсолютный путь:

```bash
export SURFACE_ATOMS_PATH="$HOME/projects/surface-atoms"
```

Рекомендуется добавить в `~/.bashrc`:

```bash
export SURFACE_ATOMS_PATH="$HOME/projects/surface-atoms"
```

---

## 4. Запуск в режиме разработки

Из каталога Studio **одной командой** (API **3847** + Vite **5173**):

```bash
cd ~/projects/surface-kinetics-studio
export SURFACE_ATOMS_PATH="$HOME/projects/surface-atoms"   # если не в .bashrc
npm run dev
```

В браузере на той же машине: **http://localhost:5173**

| Порт | Сервис |
|------|--------|
| **5173** | React (Vite), прокси `/api` → 3847 |
| **3847** | Node API (схемы, config, `go run`, результаты) |

**Важно:** не запускайте только фронт (`npm run dev --prefix client`) без API — в консоли будет `proxy ECONNREFUSED 127.0.0.1:3847`.

Альтернатива — два терминала:

```bash
npm run dev:server   # только API
npm run dev:client   # только Vite
```

Проверка API:

```bash
curl -s http://127.0.0.1:3847/api/health
```

---

## 5. Доступ с Windows (Hyper-V) через nginx

На Linux-ВМ, где запущена Studio.

**5.1.** Запустить Studio (шаг 4).

**5.2.** Один раз подключить nginx (нужен `sudo`; конфиг лежит в отдельном каталоге `nginx/` рядом с проектами, если вы его клонировали, или скопируйте `surface-kinetics-studio.conf` вручную):

```bash
sudo /path/to/nginx/enable-surface-kinetics.sh
```

Пример для нашей ВМ:

```bash
sudo ~/projects/nginx/enable-surface-kinetics.sh
```

**5.3.** Узнать IP ВМ:

```bash
hostname -I | awk '{print $1}'
```

**5.4.** С Windows в браузере:

```text
http://<IP-ВМ>:8085
```

Пример: `http://192.168.130.100:8085`

Если страница не открывается: ping до IP, фаервол Windows; на ВМ при UFW:

```bash
sudo ufw allow 8085/tcp
```

**502 Bad Gateway** — обычно не запущен `npm run dev` или Vite не слушает `5173`.

Подробнее: `nginx/README-surface-kinetics.md` (если каталог `nginx` есть в `~/projects`).

---

## 6. Production (без hot-reload)

Один процесс, порт **3847** (статика + API):

```bash
cd ~/projects/surface-kinetics-studio
export SURFACE_ATOMS_PATH="$HOME/projects/surface-atoms"

cd client && npm run build && cd ..
NODE_ENV=production npm start
```

Открыть: **http://localhost:3847**

Текущий nginx-конфиг рассчитан на **режим разработки** (`npm run dev` + Vite). Для production за nginx нужен отдельный `location` на `127.0.0.1:3847` или прямой доступ к `:3847`.

---

## 7. Переменные окружения

| Переменная | Где | По умолчанию | Назначение |
|------------|-----|--------------|------------|
| `SURFACE_ATOMS_PATH` | Studio | `../surface-atoms` | Абсолютный путь к каталогу симулятора |
| `PORT` | Studio API | `3847` | Порт бэкенда |
| `NODE_ENV` | Studio | — | `production` для `npm start` |

---

## 8. Проверка после развёртывания

1. `curl http://127.0.0.1:3847/api/health` — в ответе корректный путь к `surface-atoms`.
2. Studio → **Схема** — открывается `scheme_marinov.yaml`, сохранение без ошибок.
3. **Конфигурация** — читается/пишется `config.yaml`.
4. **Запуск** → **Собрать** — успешный `go build`.
5. **Запустить симуляцию** — в `surface-atoms` появляется новая папка `result …`.
6. **Результаты** — список прогонов, графики, Excel.

---

## 9. Типичные проблемы

| Симптом | Причина | Решение |
|---------|---------|---------|
| `proxy ECONNREFUSED 127.0.0.1:3847` | API не запущен | `npm run dev` или `npm run dev:server` |
| Неверные файлы схемы/config | Неверный `SURFACE_ATOMS_PATH` | `export` на правильный каталог, перезапустить Studio |
| `go: command not found` | Go не установлен | Установить Go 1.23+ |
| 502 через nginx | Vite не работает | `npm run dev` на ВМ, порт 5173 |
| Пустые результаты | Симуляция не завершилась | Лог в терминале, где запущен `npm run dev` |
| `missing go.sum entry for module…` | Старый клон без `go.sum` в репозитории | `git pull` (ветка `dev`) или `go mod tidy` && `go build` |

---

## 10. Рабочий цикл

1. Запустить Studio: `npm run dev`.
2. **Конфигурация** — элементы N / N+O, константы, пресет из базы → **Сохранить**.
3. **Схема** — rates / events → **Сохранить** → **Подключить в config.yaml**.
4. **Запуск** — `go build`, одиночный прогон или sweep по T и `agDensity`.
5. **Результаты** — графики и Excel из `surface-atoms/result …`.

Журналы sweep: `surface-kinetics-studio/data/sweeps/` (локально, в git не попадают).

---

## 11. Обновление с GitHub

```bash
cd ~/projects/surface-atoms && git pull
cd ~/projects/surface-kinetics-studio && git pull
cd ~/projects/surface-kinetics-studio && npm install && cd client && npm install && cd ..
```

После `git pull` с новыми API-роутами **перезапустите** `npm run dev`.

---

## Шпаргалка (копировать целиком)

```bash
mkdir -p ~/projects && cd ~/projects
git clone git@github.com:Kizerfifas/surface-atoms.git
git clone git@github.com:Kizerfifas/surface-kinetics-studio.git

cd ~/projects/surface-atoms
go mod download && go build -o surface-atoms .

cd ~/projects/surface-kinetics-studio
npm install && cd client && npm install && cd ..
export SURFACE_ATOMS_PATH="$HOME/projects/surface-atoms"
npm run dev
# → http://localhost:5173

# Опционально: nginx для доступа с Windows
# sudo ~/projects/nginx/enable-surface-kinetics.sh
# → http://<IP-ВМ>:8085
```

---

См. также: [README](../README.md) (пользовательская инструкция и описание UI).
