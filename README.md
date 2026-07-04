# Auto Platform

Платформа объявлений о продаже автомобилей, реализованная как набор независимых микросервисов на Go. Пользователи могут регистрироваться, публиковать объявления, общаться с продавцами в чате и загружать фотографии через S3-совместимое хранилище.

## Архитектура

<img width="1052" height="794" alt="изображение" src="https://github.com/user-attachments/assets/583c31d2-78ca-4da1-843f-b066a3c34d8d" />

### Архитектурные паттерны

**API Gateway** — ingress-nginx выступает единой точкой входа для всех клиентских запросов. Вместо того чтобы обращаться к каждому сервису напрямую, клиент всегда идёт на один адрес (`auto-platfrom.ru`), а шлюз сам маршрутизирует трафик по префиксу пути (`/api/auth` → auth-service, `/api/listings` → listing-service и т.д.). Помимо маршрутизации шлюз берёт на себя сквозные задачи: TLS-терминацию и проверку JWT. Последнее реализовано через директиву `auth_request` — перед каждым запросом к защищённому пути nginx делает суб-запрос к `auth-service /api/auth/validate` и форвардит запрос дальше только при ответе `200 OK`, пробрасывая `X-User-Id` в заголовке. Ни один из бизнес-сервисов не занимается декодированием JWT самостоятельно.

**Fan-out** — паттерн используется в messenger-service для доставки сообщений по WebSocket при нескольких репликах. Когда пользователь отправляет сообщение, оно сохраняется в Postgres и публикуется в Kafka-топик `messenger.message.sent`. Проблема в том, что при 3 репликах неизвестно, к какой из них подключён получатель по WebSocket. Если использовать один `group.id` на все реплики, Kafka отдаст сообщение только одной из них — и если получатель сидит на другой, доставка не произойдёт. Решение: каждая реплика при старте генерирует уникальный `group.id = "messenger.fanout." + UUID`, тем самым становясь отдельным потребителем в глазах Kafka. Все реплики получают копию каждого события и проверяют свой WebSocket-хаб — та, у которой есть соединение с получателем, доставляет сообщение, остальные игнорируют.

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

Каждый пуш в `main` (и каждый PR) проходит через следующую цепочку:

**Шаг 1 — detect-changes**
`dorny/paths-filter` анализирует diff коммита и выставляет булевы флаги (`auth: true/false`, `listing: true/false` и т.д.). Все последующие джобы читают эти флаги через `needs.detect-changes.outputs.*` и пропускаются (`skipped`), если их сервис не изменился. Изменения в `proto/` помечают `listing` и `messenger` — оба зависят от Protobuf-контрактов.

**Шаг 2 — сборка каждого изменённого сервиса (параллельно)**

Для Go-сервисов (`auth`, `listing`, `user`, `messenger`, `storage`):
1. `go build ./...` — компилирует весь сервис; если код не компилируется, пайплайн падает здесь
2. `go test ./... -cover` — запускает юнит-тесты с подсчётом покрытия
3. `golangci-lint` — агрегатор статических анализаторов; проверяет стиль, возможные баги, неиспользуемые переменные, теневые ошибки и др.
4. `docker build` — собирает образ с тегом `github.sha`
5. `docker push` → GHCR (только при push в `main`, не на PR)

> `storage-service` — единственный сервис без `librdkafka`: он не использует Kafka, поэтому CGO-зависимость не нужна и образ собирается с `CGO_ENABLED=0`.

Для `frontend`:
1. `npm ci` — детерминированная установка зависимостей из `package-lock.json`
2. `npm run build` — TypeScript-проверка (`tsc -b`) + сборка через Vite; ошибки типов ломают пайплайн
3. `npm run lint` — ESLint по правилам проекта
4. `docker build` + `docker push` → GHCR (только при push в `main`)

**Шаг 3 — deploy (только при push в `main`)**
После завершения всех джоб запускается `deploy`. Он устанавливает `yq` и обновляет `helm/auto-platform/values-production.yaml` — заменяет `image.tag` только у тех сервисов, чья джоба завершилась со статусом `success`. Упавшие и пропущенные сервисы не трогаются — в кластер уходит последний рабочий образ. Коммит делается с суффиксом `[skip ci]`, чтобы не запускать CI повторно.

### JWT (access + refresh)
Stateless аутентификация: access-токен живёт 15 минут, refresh — 7 дней. ingress-nginx проверяет каждый запрос к защищённым путям через `auth_request` к auth-service — авторизация централизована в одном месте и не дублируется в каждом сервисе.

### Swagger / OpenAPI
Документация генерируется из аннотаций в коде (`swaggo/swag`) — она всегда актуальна, потому что живёт рядом с хендлерами.

---

## Что я узнал в процессе разработки

Проект строился не по учебнику — каждый блок приносил проблемы, которые пришлось решать самостоятельно. Ниже — самые ценные из них.

**Логгер падал при пустой строке пути.**
Когда `LOGGER_FOLDER` был не задан, `os.MkdirAll("")` возвращал ошибку и сервис уходил в `CrashLoopBackOff`. Поверхностный взгляд на код не выявил бы проблему — пришлось разобраться как Go обрабатывает пустую строку в файловых операциях и добавить явную ветку для stdout-only режима.

**Kafka ACL — тема и группа это разные объекты.**
Сервис подключался к брокеру, топик читался, но при старте консьюмера прилетало `Group authorization failed`. Оказалось, что в Kafka ACL на топик и ACL на consumer group — независимые разрешения. Managed-сервис Timeweb не давал управлять группами через UI, пришлось разбираться с `kafka-acls.sh`, форматом JAAS-конфигурации и отлаживать аутентификацию SCRAM-SHA-512 вручную.

**PostgreSQL в managed-сервисе: имя пользователя имеет значение.**
Имена с префиксом `pg_` зарезервированы PostgreSQL для системных ролей — `initdb` отказывался создавать такого пользователя при поднятии self-hosted инстанса внутри кластера. Это не очевидно из документации и понимается только при чтении исходников initdb.

**GitOps: деплой — это коммит, а не команда.**
Первое время я запускал `helm upgrade` вручную. После перехода на ArgoCD пришлось полностью переосмыслить процесс: CI коммитит новый image tag в git, ArgoCD подхватывает изменение и синхронизирует кластер. Откат — это `git revert`. Это меняет мышление с императивного («применить») на декларативное («желаемое состояние»).

**Несколько values-файлов в Helm: порядок решает.**
Чарт поддерживает три слоя конфигурации (`values.yaml` → `values-production.yaml` → `values-production-self-hosted.yaml`). Позднее понял, что Helm мержит вложенные объекты по ключу, а не заменяет целиком — из-за этого `external: false` в последнем файле корректно перекрывал `external: true` из предыдущего, но только если оба файла переданы ArgoCD. Один пропущенный файл — и поды уходят на старый IP.

**PVC не создаются без StorageClass.**
В bare-metal кластере нет динамического провижининга по умолчанию. `StatefulSet` зависал в `Pending` без каких-либо внятных ошибок. Решение — установить `local-path-provisioner` и пометить его как дефолтный `StorageClass`.

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
