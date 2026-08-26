# terraria-mod-update-automatization

CLI-утилита на Go для автоматической проверки обновлений модов tModLoader в Steam Workshop и их синхронизации с Git-репозиторием сервера Terraria.

Проект разрабатывался для автоматизации сборки и поставки модов под сервер [f3rym/terraria-server](https://github.com/f3rym/terraria-server).

---

### Как это работает

1. Yаходит все `.tmod` файлы в локальной директории Steam Workshop (`1281930` - директория TModLoader).
2. Делает батч-запрос `GetPublishedFileDetails` и получает `time_updated` для каждого мода.
3. Сравнивает timestamps с локальным `mods_cache_dirijable.json`. Если кэша нет или версия в Steam новее - мод помечается на обновление.
4. **Деплой**:
* Выполняет `git fetch origin && git reset --hard origin/main` в целевом репозитории сервера, чтобы избежать конфликтов.
* Копирует обновлённые `.tmod` файлы.
* Обновляет JSON-кэш.
* Делает `git add`, `git commit` и `git push origin main`.

---

### Сборка и запуск

#### Сборка бинарника:

```bash
go build -o mod_sync.exe .

```

#### Параметры запуска:

| Флаг | Описание |
| --- | --- |
| `-sp` (Steam Path) | Путь к директории модов Steam Workshop (`.../steamapps/workshop/content/1281930`) |
| `-pp` (Postavka Path) | Путь к локальной копии репозитория сервера Terraria |

#### Пример вызова:

```bash
./mod_sync.exe \
  -sp "D:\Steam\steamapps\workshop\content\1281930" \
  -pp "C:\DevOps\terraria-server"
```