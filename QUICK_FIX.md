# Быстрое решение для Windows

## ❌ Проблема:
```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work
```

## ✅ Решение (выберите одно):

### 1. Docker (самый простой) ⭐

```powershell
cd ..
docker-compose up --build
```

**Преимущества:**
- Не нужно устанавливать дополнительные программы
- Всё работает "из коробки"
- Уже установлен у вас

---

### 2. WSL (рекомендуется для разработки)

1. Установите WSL: `wsl --install` в PowerShell (как администратор)
2. Перезагрузите компьютер
3. В WSL:
   ```bash
   sudo apt update && sudo apt install gcc -y
   cd /mnt/c/Users/Lenovo/Downloads/go-whatsapp-web-multidevice-main/go-whatsapp-web-multidevice-main/src
   go run . rest
   ```

---

### 3. MinGW (для нативной Windows)

**Если хотите запускать напрямую в Windows:**

1. Скачайте TDM-GCC: https://jmeubank.github.io/tdm-gcc/download/
2. Установите (обычно в `C:\TDM-GCC-64`)
3. Добавьте в PATH: `C:\TDM-GCC-64\bin`
4. Перезапустите PowerShell
5. Проверьте: `gcc --version`
6. Запустите:
   ```powershell
   $env:CGO_ENABLED = "1"
   cd src
   go run . rest
   ```

---

## 🎯 Мой совет:

**Используйте Docker** - это самое быстрое и простое решение!

```powershell
cd ..
docker-compose up --build
```

