# WB_project_2.12


Для компиляции утилиты выполните команду в этой директории:
```
make mygrep
```
Для запуска тестов:
```
bash test.sh
```
Можно запускать вручную через `./mygrep` и так же как обычный `grep`.


Я сократил количество тестов, так как тесты пишутся дольше чем сама утилита, их количество в sort было излишне.
Для проверки кода выполните следующие команды:
```
golangci-lint run ./...
go vet ./...
```