# Настройка для Windows

## Проблема: CGO требуется для go-sqlite3

Этот проект использует `go-sqlite3`, который требует CGO. На Windows для CGO нужен GCC компилятор.

## Решения:

### Вариант 1: Docker (РЕКОМЕНДУЕТСЯ - Самый простой) ✅

Используйте Docker - он уже работает у вас:

```powershell
cd ..
docker-compose up --build
```

Откройте: `http://localhost:3000`

---

### Вариант 2: WSL (Windows Subsystem for Linux) ✅

1. Установите WSL: https://docs.microsoft.com/en-us/windows/wsl/install
2. В WSL терминале:
   ```bash
   sudo apt update
   sudo apt install gcc
   cd /mnt/c/Users/Lenovo/Downloads/go-whatsapp-web-multidevice-main/go-whatsapp-web-multidevice-main/src
   go run . rest
   ```

---

### Вариант 3: Установить MinGW-w64 для Windows

1. **Установите MinGW-w64:**
   - Скачайте: https://www.mingw-w64.org/downloads/
   - Или через MSYS2: https://www.msys2.org/
   - Или используйте TDM-GCC: https://jmeubank.github.io/tdm-gcc/

2. **Добавьте GCC в PATH:**
   - Добавьте путь к `bin` папке MinGW в переменную PATH
   - Например: `C:\mingw64\bin` или `C:\TDM-GCC-64\bin`

3. **Установите переменную окружения:**
   ```powershell
   $env:CGO_ENABLED = "1"
   ```

4. **Проверьте:**
   ```powershell
   gcc --version
   go env CGO_ENABLED
   ```

5. **Запустите:**
   ```powershell
   cd src
   $env:CGO_ENABLED = "1"
   go run . rest
   ```

---

## Быстрое решение прямо сейчас:

Проще всего использовать **Docker**:

```powershell
# Остановите текущий процесс (Ctrl+C если запущен)
cd ..
docker-compose up --build
```

Или в фоне:
```powershell
docker-compose up -d --build
```

Затем откройте: `http://localhost:3000`

