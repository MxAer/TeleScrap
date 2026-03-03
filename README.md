# TeleScrap

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791)](https://www.postgresql.org/)
 [![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-green.svg)](https://www.gnu.org/licenses/agpl-3.0)

Скраппер-юзербот для Telegram на Go с интерфейсом для просмотра собранных данных


![alt text](docs/example.png) 

## Функционал

- Сбор сообщений, а так же файлов из каналов или личных сообщений
  
- Сбор информации о профилях пользователей
  
- Просмотр собранных данных через веб-интерфейс
  

## Запуск

Для запуска необходим [Docker](https://docs.docker.com/engine/install/)

- Склонируйте репозиторий с помощью `git clone https://github.com/MxAer/TeleScrap`
  

- Создайте файл со своими переменными для запуска по примеру [.env.example](https://github.com/MxAer/TeleScrap/.env.example)
  

- Введите в терминал, открытый в папке склонированного репозитория `docker compose up -d`
  

- Вы успешно запустили TeleScrap! Веб-интерфейс вы можете найти на порте 6767
