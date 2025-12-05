#!/bin/bash

# Запускаем инфраструктуру в фоне без логов
docker-compose up 2>&1 | grep -v -E "(redis|zookeeper|kafka|clickhouse)"