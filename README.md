# Auto Platform

Платформа объявлений о продаже автомобилей, реализованная как набор независимых микросервисов на Go. Пользователи могут регистрироваться, публиковать объявления, общаться с продавцами в чате и загружать фотографии через S3-совместимое хранилище.

## Архитектура

```
                        ┌─────────────────────────────────────────────────┐
                        │                 ingress-nginx                   │
                        │   /api/auth  /api/listings  /api/user  ...      │
                        │          auth_request → auth-service            │
                        └─────────┬───────┬───────┬───────┬──────────────┘
                                  │       │       │       │
              ┌───────────────────┼───────┼───────┼───────┼──────────────┐
              │                   ▼       ▼       ▼       ▼              │
              │  ┌────────────┐ ┌──────┐ ┌──────┐ ┌─────────┐ ┌───────┐│
              │  │auth-service│ │list- │ │user- │ │messenger│ │storage││
              │  │ JWT/bcrypt │ │ing   │ │serv. │ │-service │ │-serv. ││
              │  └─────┬──────┘ └──┬───┘ └──┬───┘ └────┬────┘ └───────┘│
              │        │           │  gRPC   │           │               │
              │        │           └─────────┘           │               │
              │        │                                 │               │
              │   ┌────▼────┐  ┌──────────┐  ┌─────────▼──┐            │
              │   │Postgres │  │  Redis   │  │  Postgres  │            │
              │   │(auth_db)│  │  cache   │  │(messenger) │            │
              │   └─────────┘  └──────────┘  └────────────┘            │
              │                                                          │
              │              Apache Kafka (async events)                 │
              │         user.register → user-service consumer            │
              └──────────────────────────────────────────────────────────┘
```

### Сервисы

| Сервис | Порт | Назначение |
|---|---|---|
| **auth-service** | 5050 | Регистрация, вход, JWT access/refresh токены, валидация для ingress |
| **listing-service** | 6060 / 6061 (gRPC) | CRUD объявлений, кэш в Redis, gRPC для межсервисных вызовов |
| **user-service** | 5050 | Профили пользователей, консьюмер Kafka-событий регистрации |
| **messenger-service** | 7070 | Чат между покупателем и продавцом, WebSocket |
| **storage-service** | 7080 | Presigned URL для загрузки/скачивания файлов через S3 |

## Стек технологий и обоснование выбора

### Go
Микросервисы написаны на Go потому что язык изначально проектировался для сетевых серверов: горутины дешевле потоков ОС на порядок, стандартная библиотека закрывает большинство HTTP/gRPC задач, а бинарники компилируются в один статический файл — это упрощает Docker-образы до `FROM scratch`. Статическая типизация и явная обработка ошибок снижают вероятность runtime-паник по сравнению с динамическими языками.

### Gin
HTTP-фреймворк с роутером на radix-дереве и минимальными аллокациями. Выбран за то, что не скрывает стандартный `net/http` — при необходимости можно выйти на уровень `http.Handler` без переписывания.

### PostgreSQL
Реляционная БД с ACID-транзакциями. Каждый сервис имеет собственную схему (Database-per-Service) — это граница bounded context и единственный способ менять схему одного сервиса без координации с командами других. Для клиента выбран `pgx/v5` вместо `database/sql` — он поддерживает нативный протокол PostgreSQL и именованные параметры.

### Apache Kafka
Асинхронная шина событий. При регистрации auth-service публикует событие `user.register` и сразу отвечает клиенту — user-service дочитывает событие независимо. Это развязывает сервисы: падение user-service не ломает регистрацию, сообщение будет обработано позже. Kafka выбрана вместо RabbitMQ из-за гарантии сохранности сообщений на диске и возможности повтора с произвольного offset.

### Redis
Кэш объявлений в listing-service. Популярные запросы (поиск, главная страница) не идут в PostgreSQL — ответ отдаётся из памяти за < 1 мс. Выбран вместо Memcached за поддержку структур данных (Hash, Sorted Set) и pub/sub.

### gRPC + Protobuf
Messenger-service узнаёт продавца объявления при создании треда — синхронный внутренний вызов в listing-service по gRPC. По сравнению с REST: строгая схема (`.proto`), бинарная сериализация (меньше байт чем JSON), кодогенерация клиента и сервера устраняет рассинхронизацию контрактов.

### Kubernetes + Helm
Оркестратор с декларативным управлением. Helm-чарт параметризует весь стек одним файлом `values.yaml`: можно поднять dev-окружение в minikube и production в облаке одной командой, меняя только значения. `replicaCount: 3` для каждого сервиса даёт горизонтальное масштабирование без изменений кода.

### ArgoCD (GitOps)
ArgoCD следит за веткой `main` и синхронизирует кластер с тем, что лежит в git. CI после успешного билда коммитит новый image tag в `values-production.yaml` — это единственный «деплой». Откат — это `git revert`. Аудит изменений — это `git log`.

### GitHub Actions CI
Монорепозиторий с детектированием изменений (`dorny/paths-filter`): если изменился только `listing-service/`, пересобирается только он. Образы публикуются в GitHub Container Registry (GHCR) с тегом по SHA коммита.

### JWT (access + refresh)
Stateless аутентификация: access-токен живёт 15 минут, refresh — 7 дней. ingress-nginx проверяет каждый запрос к защищённым путям через `auth_request` к auth-service — авторизация централизована в одном месте и не дублируется в каждом сервисе.

### Swagger / OpenAPI
Документация генерируется из аннотаций в коде (`swaggo/swag`) — она всегда актуальна, потому что живёт рядом с хендлерами.

---

## Быстрый старт — Minikube (локально)

### Требования
- [Docker](https://docs.docker.com/get-docker/)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm 3](https://helm.sh/docs/intro/install/)

### 1. Запустить minikube

```bash
minikube start 
minikube addons enable ingress
```

### 2. Установить local-path-provisioner (динамические PVC)

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.31/deploy/local-path-storage.yaml
kubectl patch storageclass local-path -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

### 3. Создать namespace и секреты

```bash
kubectl create namespace auto-platform

# JWT
kubectl create secret generic auto-platform-auth \
  --from-literal=JWT_SECRET=$(openssl rand -base64 32) \
  -n auto-platform

# PostgreSQL (один секрет на каждый сервис)
for svc in auth listing user messenger; do
  kubectl create secret generic auto-platform-${svc}-postgres \
    --from-literal=password=$(openssl rand -base64 16) \
    -n auto-platform
done

# Kafka (если используете внешний брокер; для minikube можно пропустить,
# сервисы запустятся без Kafka — только HTTP-функциональность)
kubectl create secret generic kafka-secret \
  --from-literal=username=YOUR_KAFKA_USER \
  --from-literal=password=YOUR_KAFKA_PASSWORD \
  -n auto-platform

# S3 (Timeweb Cloud Storage или любой S3-совместимый)
kubectl create secret generic auto-platform-storage \
  --from-literal=S3_ACCESS_KEY=YOUR_KEY \
  --from-literal=S3_SECRET_KEY=YOUR_SECRET \
  -n auto-platform
```

### 4. Задеплоить чарт

```bash
helm upgrade --install auto-platform ./helm/auto-platform \
  --namespace auto-platform \
  --set global.secretsManagedExternally=true \
  --set ingress.host=$(minikube ip).nip.io \
  --set global.kafkaBrokers="YOUR_BROKER:9092"
```

### 5. Открыть в браузере

```bash
# Swagger документация auth-service
curl http://$(minikube ip).nip.io/api/auth/swagger/index.html

# Проверить health всех сервисов
for svc in auth listings user messenger storage; do
  echo -n "$svc: "
  curl -s -o /dev/null -w "%{http_code}" http://$(minikube ip).nip.io/api/$svc/health
  echo
done
```

---

## Деплой в production-кластер (Timeweb Cloud / любой k8s)

### Требования
- Kubernetes кластер с установленным ingress-nginx
- ArgoCD в namespace `argocd`
- Форк этого репозитория (ArgoCD смотрит на ваш git)

### 1. Установить ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### 2. Создать секрет для pull образов из GHCR

```bash
kubectl create secret docker-registry ghcr-pull-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_PAT \
  --docker-email=your@email.com \
  -n auto-platform
```

### 3. Создать секреты приложения

Аналогично шагу 3 из minikube-инструкции, добавив kafka-secret с реальными данными брокера.

### 4. Применить ArgoCD Application

Поменяйте `repoURL` в `argocd/application.yaml` на ваш форк, затем:

```bash
kubectl apply -f argocd/application.yaml
```

ArgoCD начнёт следить за веткой `main` и задеплоит чарт. Все последующие изменения применяются через `git push`.

### 5. Настроить DNS

Направьте A-запись вашего домена на внешний IP ingress-nginx:

```bash
kubectl get svc -n ingress-nginx ingress-nginx-controller
```

### CI/CD

После настройки репозитория добавьте в GitHub Secrets:
- `GH_PAT` — Personal Access Token с правом `write:packages` (для публикации образов в GHCR)

При каждом `push` в `main` GitHub Actions:
1. Определяет какие сервисы изменились
2. Собирает только их Docker-образы
3. Публикует в GHCR с тегом по SHA коммита
4. Коммитит новый тег в `values-production.yaml`
5. ArgoCD обнаруживает изменение и деплоит

---

## Структура репозитория

```
auto-platform/
├── auth-service/          # Аутентификация и авторизация
├── listing-service/       # Объявления (HTTP + gRPC сервер)
├── user-service/          # Профили пользователей
├── messenger-service/     # Чат (HTTP + WebSocket)
├── storage-service/       # Presigned S3 URL
├── proto/                 # Protobuf-контракты (listing gRPC)
├── helm/
│   └── auto-platform/     # Helm-чарт всего стека
│       ├── values.yaml                        # Дефолты
│       ├── values-production.yaml             # Production оверрай
│       └── values-production-self-hosted.yaml # Postgres внутри кластера
├── argocd/
│   └── application.yaml   # GitOps — применяется один раз вручную
└── .github/
    └── workflows/
        └── ci.yml          # Сборка, тесты, публикация образов
```

## Структура каждого сервиса

```
<service>/
├── cmd/
│   └── main.go            # Точка входа, инициализация зависимостей
├── internal/
│   ├── core/
│   │   ├── config/        # Конфигурация через envconfig
│   │   ├── domain/        # Сущности и интерфейсы репозиториев
│   │   ├── logger/        # Zap-логгер (stdout + файл)
│   │   └── transport/     # Kafka producer/consumer, gRPC клиент
│   └── features/
│       └── <feature>/
│           ├── repository/ # SQL-запросы (pgx)
│           ├── service/    # Бизнес-логика
│           └── transport/  # HTTP-хендлеры (Gin)
├── docs/                  # Swagger (генерируется swag init)
├── Dockerfile
└── go.mod
```

Архитектура следует принципу Dependency Inversion: `service` зависит от интерфейсов `domain`, а не от конкретных `repository` или `transport`. Это позволяет тестировать бизнес-логику без поднятия БД.

## API

После запуска Swagger UI доступен по адресам:

| Сервис | URL |
|---|---|
| auth-service | `http://<host>/api/auth/swagger/index.html` |
| listing-service | `http://<host>/api/listings/swagger/index.html` |
| user-service | `http://<host>/api/user/swagger/index.html` |
| messenger-service | `http://<host>/api/messenger/swagger/index.html` |
| storage-service | `http://<host>/api/storage/swagger/index.html` |

### Основные эндпоинты

```
POST   /api/auth/register          Регистрация
POST   /api/auth/login             Вход, выдача токенов
POST   /api/auth/refresh           Обновление access-токена
GET    /api/auth/authorized        Валидация токена (для ingress auth_request)

GET    /api/listings               Список объявлений
POST   /api/listings               Создать объявление [требует JWT]
GET    /api/listings/{id}          Одно объявление
GET    /api/listings/mine          Мои объявления [требует JWT]

GET    /api/user/me                Мой профиль [требует JWT]

GET    /api/messenger/mine         Мои треды [требует JWT]
GET    /api/messenger/ws           WebSocket-подключение [требует JWT]

POST   /api/storage/upload         Загрузить файл [требует JWT]
```
