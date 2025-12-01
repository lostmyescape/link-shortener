#!/bin/bash

# Запускаем инфраструктуру в фоне без логов
docker-compose up -d zookeeper kafka

# Ждем пока Kafka будет готова
sleep 10

# Запускаем только наши сервисы и фильтруем логи
docker-compose up --build analytics auth shortener 2>&1 | \
  grep -v "kafka" | \
  grep -v "zookeeper" | \
  grep -v "kafdrop"