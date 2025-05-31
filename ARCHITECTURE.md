
# Архитектура микросервисного приложения (соцсеть рецептов)

Общая архитектура приложения спроектирована по принципам микросервисного подхода: каждый компонент (аутентификация, управление пользователями, рецепты, чат, медиа, нотификации, поиск, аналитика) реализован отдельным сервисом. Все сервисы написаны на Go и структурированы по слоям «контроллер – бизнес-логика – репозиторий – клиент БД/клиент очереди/клиент кеша». Клиентские приложения (веб/мобильные) обращаются к единой точке входа – API Gateway, который выполняет маршрутизацию, аутентификацию и другие общие задачи.

Межсервисное взаимодействие организовано как синхронно, так и асинхронно:
- **Синхронные запросы (REST/gRPC).** Многие операции (например, получение профиля пользователя или списка рецептов) происходят через HTTP/JSON API (через Gin/Echo) либо через gRPC (сгенерированный Go-клиент). gRPC обеспечивает эффективную бинарную сериализацию данных и поддержку двунаправленного стриминга, что важно для высокопроизводительного обмена между микросервисами (например, в чате).
- **Асинхронные сообщения (Kafka).** Для обмена событиями и фоновых задач используется брокер сообщений Apache Kafka. Это позволяет публиковать события `UserRegistered`, `RecipeCreated`, `MessageSent` и другие, на которые подписываются нужные сервисы. Например, при регистрации пользователя Auth-сервис публикует событие `UserRegistered` в Kafka, и User-сервис создаёт профиль, а Notification-сервис отправляет приветственное письмо. При создании рецепта Recipe-сервис публикует событие `RecipeCreated`, на которое подписаны Search-сервис (индексация), Analytics (статистика) и Notification (уведомление подписчиков).
- **Кеширование и ускорение запросов.** Для быстрых ответов на часто запрашиваемые данные используется Redis (через go-redis). Например, кешируются популярные рецепты, сессии пользователей, результаты сложных запросов. Redis также может использоваться как брокер pub/sub для уведомлений в реальном времени.
- **Мониторинг и трассировка.** Для сбора метрик и логов каждый сервис интегрируется с Prometheus/Grafana и Jaeger (или Zipkin), а также проводит логирование через zap или logrus.

В целом, такая архитектура обеспечивает гибкое масштабирование каждого сервиса, отказоустойчивость и легкость расширения. Использование **API Gateway** скрывает сложность множества эндпоинтов и централизует кросс-сервисные задачи (аутентификация, агрегация данных).

---

## Структура репозитория (Monorepo)

```
monorepo/
├── api-gateway/           # Код и конфигурация API-шлюза
├── auth-service/          # Сервис аутентификации и управления аккаунтами
├── user-service/          # Сервис профилей пользователей и соц. связей
├── recipe-service/        # Сервис управления рецептами
├── chat-service/          # Сервис личных сообщений
├── media-service/         # Сервис загрузки/хранения медиа-файлов
├── notification-service/  # Сервис уведомлений (push/email)
├── search-service/        # Сервис поисковой индексации (Elasticsearch)
├── analytics-service/     # Сервис сбора аналитики и событий
├── common/                # Общие библиотеки, DTO, утилиты
├── deployment/            # Манифесты инфраструктуры (Docker Compose, Kubernetes, Terraform)
└── docs/                  # Документация проекта
```

Каждый сервис содержит собственные исходники, Dockerfile, скрипты миграций БД и т. д. Папка `common/` включает код, используемый несколькими сервисами (модели, клиенты, утилиты).

---

## Схемы баз данных (PostgreSQL)

### Auth Service

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
```

### User Service

```sql
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL,
    full_name VARCHAR(100),
    bio TEXT,
    avatar_media_id INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE followers (
    user_id INTEGER NOT NULL,
    follower_id INTEGER NOT NULL,
    followed_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, follower_id)
);
```

### Recipe Service

```sql
CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    author_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    ingredients TEXT,
    steps TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE recipe_images (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    media_id INTEGER NOT NULL,
    position INT,
    description TEXT
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE likes (
    recipe_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    liked_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (recipe_id, user_id)
);
```

### Chat Service

```sql
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    is_group BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE conversation_members (
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT NOW()
);
```

### Media Service

```sql
CREATE TABLE media (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL,
    file_type VARCHAR(50),
    url TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);
```

### Notification Service

```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    type VARCHAR(50),
    reference_id INTEGER,
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Analytics Service

```sql
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    event_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,
    properties JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

> Search Service использует Elasticsearch; отдельной SQL-схемы не требует (данные индексируются из Kafka).

---

## Микросервисы: внутреннее строение и технологии

### 1. Auth Service  
**Назначение:** Регистрация, аутентификация, выдача JWT, refresh-токены.

- **Контроллер (Handler):**  
  - Роуты через Gin:  
    - `POST /auth/register` – регистрация  
    - `POST /auth/login` – вход, выдача JWT  
    - `POST /auth/refresh` – обновление токена  
    - `GET /auth/logout` – выход  
    - `GET /auth/validate` – проверка токена  
- **Сервис (Use Case):**  
  - Проверка уникальности email/username  
  - Хеширование пароля (bcrypt)  
  - Генерация/валидация JWT (`github.com/golang-jwt/jwt/v4`)  
  - Управление refresh-токенами  
- **Репозиторий (Repository):**  
  - PostgreSQL через GORM или `pgx/SQLx`  
  - Таблицы: `users`, `refresh_tokens`  
- **Клиенты/Внешние зависимости:**  
  - Redis (go-redis) – хранение черного списка токенов или сессий  
  - Kafka (segmentio/kafka-go) – публикация `UserRegistered`  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **БД:** GORM или `pgx/SQLx`  
  - **Кеш:** go-redis  
  - **Авторизация:** bcrypt (`golang.org/x/crypto/bcrypt`), JWT (`github.com/golang-jwt/jwt/v4`)  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap или logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 2. User Service  
**Назначение:** Управление профилями пользователей, подписки/подписчики.

- **Контроллер (Handler):**  
  - `GET /users/{id}` – получить профиль  
  - `PUT /users/{id}` – обновить профиль (только пользователь)  
  - `GET /users/{id}/followers` – список подписчиков  
  - `GET /users/{id}/following` – список подписок  
  - `POST /users/{id}/follow` – подписаться  
  - `POST /users/{id}/unfollow` – отписаться  
- **Сервис (Use Case):**  
  - Проверка прав (JWT → user_id)  
  - Логика подписки (обновление таблицы `followers`)  
  - Обработка изображений профиля (Media-сервис)  
- **Репозиторий (Repository):**  
  - PostgreSQL через GORM или `pgx/SQLx`  
  - Таблицы: `user_profiles`, `followers`  
- **Клиенты/Внешние зависимости:**  
  - Kafka – подписка на `UserRegistered` (для создания профиля), публикация `UserFollowed`/`UserUnfollowed`  
  - Redis – кеширование профилей (TTL 60s)  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **БД:** GORM или `pgx/SQLx`  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 3. Recipe Service  
**Назначение:** CRUD операций с рецептами, комментарии, лайки.

- **Контроллер (Handler):**  
  - `POST /recipes` – создать рецепт  
  - `GET /recipes/{id}` – получить рецепт  
  - `PUT /recipes/{id}` – обновить рецепт (автор)  
  - `DELETE /recipes/{id}` – удалить рецепт (автор)  
  - `GET /users/{id}/recipes` – список рецептов автора  
  - `GET /recipes?tag={tag}&page={n}` – поиск/фильтрация по тегам  
- **Сервис (Use Case):**  
  - Валидация входных данных  
  - Связь с Media (загрузка изображений → делегировать Media-сервису)  
  - Управление комментариями и лайками  
  - Публикация событий (Kafka)  
- **Репозиторий (Repository):**  
  - PostgreSQL через GORM или `pgx/SQLx`  
  - Таблицы: `recipes`, `recipe_images`, `comments`, `likes`  
- **Клиенты/Внешние зависимости:**  
  - Kafka – публикация `RecipeCreated`, `RecipeUpdated`, `RecipeDeleted`  
  - Redis – кеширование популярных/часто запрашиваемых рецептов  
  - Media-сервис – получение URL изображений  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **БД:** GORM или `pgx/SQLx`  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 4. Chat Service  
**Назначение:** Реальный обмен сообщениями (One-to-One и групповые).

- **Контроллер (Handler):**  
  - `GET /chat/rooms` – список чатов пользователя  
  - `POST /chat/rooms` – создать чат (1:1 или группа)  
  - `GET /chat/rooms/{id}/messages` – получить историю чата  
  - `POST /chat/rooms/{id}/messages` – отправить сообщение  
  - WebSocket: `/ws/chat/{room_id}` – реальный обмен сообщениями  
- **Сервис (Use Case):**  
  - Управление участниками чата  
  - Проверка доступа (только участник может писать)  
  - Бизнес-логика сообщений (сохранение, стриминг)  
- **Repository (Repository):**  
  - PostgreSQL или MongoDB  
  - Таблицы/коллекции: `conversations`, `conversation_members`, `messages`  
  - Redis/NATS – pub/sub для рассылки сообщений участникам  
- **Клиенты/Внешние зависимости:**  
  - Kafka – публикация `MessageSent`, подписка для Analytics и Notification  
  - Redis – хранение онлайна/статуса участников (для real-time)  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **WebSocket:** Gorilla Websocket или gRPC Stream  
  - **БД:** GORM (PostgreSQL) или MongoDB Driver  
  - **Pub/Sub:** go-redis (Redis) или NATS (`nats.go`)  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 5. Media Service  
**Назначение:** Загрузка, обработка и предоставление медиа (изображения, видео).

- **Контроллер (Handler):**  
  - `POST /media/upload` – загрузить файл (multipart/form-data)  
  - `GET /media/{id}` – получить информацию о файле (URL)  
  - `DELETE /media/{id}` – удалить  
  - `GET /media/user/{id}` – список медиа по пользователю  
- **Сервис (Use Case):**  
  - Валидация формата файла (изображение/видео)  
  - Обработка изображений (генерация превью, ресайз)  
  - Подготовка/конвертация видео (FFmpeg через отдельный воркер)  
  - Сохранение метаданных в БД  
- **Repository (Repository):**  
  - PostgreSQL или MongoDB (метаданные)  
  - Таблица/коллекция: `media`  
- **Хранение файлов:**  
  - MinIO или AWS S3 (через AWS SDK или `minio-go`)  
  - Для изображений – `bimg` или `github.com/disintegration/imaging`  
  - Для видео – FFmpeg воркеры (через `os/exec` или `github.com/u2takey/ffmpeg-go`)  
- **Клиенты/Внешние зависимости:**  
  - Kafka – публикация `MediaUploaded`, `MediaDeleted`  
  - Redis – кеширование путей к медиа  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **Обработка изображений:** `bimg` или `imaging`  
  - **Хранилище:** AWS SDK (S3) или `minio-go`  
  - **Видеоконвертация:** FFmpeg + `ffmpeg-go`  
  - **БД:** GORM (PostgreSQL) или MongoDB Driver  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 6. Notification Service  
**Назначение:** Обработка событий и рассылка уведомлений (push, email).

- **Контроллер (Handler):**  
  - `POST /notifications/subscribe` – регистрация устройства для push  
  - `GET /notifications/{user_id}` – получить список уведомлений  
  - `POST /notifications/send` – внутренний endpoint (для отправки)  
- **Сервис (Use Case):**  
  - Обработка подписки устройств (FCM/APNs tokens)  
  - Приём событий из Kafka (`UserFollowed`, `RecipeCommented`, `MessageSent`)  
  - Формирование и отправка уведомлений через FCM (`firebase-admin-go`), SMTP (`gomail`) или SMS API  
  - Логика таймаутов и ограничения частоты (throttling)  
- **Repository (Repository):**  
  - PostgreSQL (таблица: `notifications`) или MongoDB  
- **Клиенты/Внешние зависимости:**  
  - Kafka – потребитель `UserFollowed`, `RecipeCommented`, `MessageSent`  
  - Redis – хранение токенов и временных ограничений (rate limit)  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **Firebase:** `firebase-admin-go` (push)  
  - **Email:** `gomail` или `net/smtp`  
  - **SMS:** Twilio API (`twilio-go`)  
  - **БД:** GORM (PostgreSQL) или MongoDB Driver  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 7. Search Service  
**Назначение:** Полнотекстовый поиск по рецептам и пользователям.

- **Контроллер (Handler):**  
  - `GET /search/recipes?q={query}&page={n}` – поиск рецептов  
  - `GET /search/users?q={query}` – поиск пользователей  
  - `GET /search/tags?q={tag}` – поиск по тегам  
- **Сервис (Use Case):**  
  - Обработка запроса, формирование поиска в Elasticsearch  
  - Поддержка пагинации и фильтров  
- **Repository (Repository):**  
  - Elasticsearch (через Go-клиент `olivere/elastic`)  
  - Redis – кеширование популярных запросов  
- **Клиенты/Внешние зависимости:**  
  - Kafka – потребитель `RecipeCreated`, `RecipeUpdated`, `RecipeDeleted`, `UserUpdated` для индексирования/удаления документов  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **Поиск:** `github.com/olivere/elastic/v7` или `github.com/meilisearch/meilisearch-go`  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 8. Analytics Service  
**Назначение:** Сбор статистики, агрегация данных, предоставление метрик.

- **Контроллер (Handler):**  
  - `GET /analytics/summary` – сводная статистика (DAU, регистраций)  
  - `GET /analytics/recipes/popular` – популярные рецепты  
  - `GET /analytics/users/active` – активные пользователи  
- **Сервис (Use Case):**  
  - Потребление событий из Kafka (`UserRegistered`, `RecipeCreated`, `MessageSent`, `UserFollowed`)  
  - Агрегация данных (daily/weekly/monthly)  
  - Вычисление метрик (DAU, MAU, retention)  
- **Repository (Repository):**  
  - ClickHouse (рекомендуется) или PostgreSQL/TimescaleDB  
- **Клиенты/Внешние зависимости:**  
  - Kafka – потребитель ключевых событий  
  - Prometheus client – выставление метрик (`/metrics`)  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo  
  - **Analytics DB:** ClickHouse Go driver (`github.com/ClickHouse/clickhouse-go`) или `github.com/jackc/pgx` + TimescaleDB  
  - **Метрики:** `github.com/prometheus/client_golang/prometheus`  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

### 9. API Gateway  
**Назначение:** Шлюз, аутентификация, маршрутизация, агрегация запросов.

- **Контроллер (Handler):**  
  - Приём всех внешних запросов:  
    - `POST /auth/*`  → Auth Service  
    - `GET /users/*`  → User Service  
    - `GET /profiles/*` → (возможно геттеры профиля)  
    - `GET /recipes/*` → Recipe Service  
    - `POST /chat/rooms` → Chat Service  
    - `GET /search/*` → Search Service  
    - `POST /notifications/*` → Notification Service  
    - `GET /analytics/*` → Analytics Service  
    - `POST /media/*` → Media Service  
  - Валидирует JWT через Auth (через gRPC или публичный ключ).  
  - Контролирует rate limiting (Redis).  
- **Сервис (Use Case):**  
  - Логика маршрутизации (сопоставление пути → сервис)  
  - Агрегация (если нужно собрать ответы из нескольких сервисов)  
  - Кеширование GET-запросов (Redis)  
- **Клиенты/Внешние зависимости:**  
  - HTTP/gRPC клиенты для вызова микросервисов  
  - Redis – кеш, rate limiting  
  - Kafka – для CQRS (команды публикуются → подписчики)  
- **Технологии и библиотеки:**  
  - **HTTP:** Gin/Echo или внешние решения (Kong, Traefik)  
  - **gRPC:** `google.golang.org/grpc`, `grpc-gateway`  
  - **Кеш:** go-redis  
  - **Очереди:** segmentio/kafka-go  
  - **Логирование:** zap/logrus  
  - **Конфиг:** Viper  
  - **Dockerfile & Makefile**  

---

## План запуска проекта

1. **Инфраструктура:**  
   - Развернуть Docker Compose или Kubernetes-окружение.  
   - Запустить Zookeeper + Kafka, базы данных (PostgreSQL для каждого микросервиса, Redis, ClickHouse/TimescaleDB для аналитики), MinIO/S3, Elasticsearch.  
   - Настроить Prometheus и Grafana (и Jaeger для трассировки).  
   - Подготовить TLS-сертификаты и конфигурации сетей.  

2. **Конфигурация:**  
   - Создать конфиги (через Viper) для каждого сервиса: DB_URL, KAFKA_BROKERS, REDIS_URL, S3_CONFIG, JWT_SECRET, etc.  
   - Настроить секреты (базы, JWT, S3) через Kubernetes Secrets или Docker secrets.  

3. **Запуск Auth и User:**  
   - Выполнить миграции БД (Auth → `users`, `refresh_tokens`; User → `user_profiles`, `followers`).  
   - Запустить Auth-service, протестировать регистрацию и вход (Postman).  
   - Запустить User-service, протестировать создание/обновление профиля; проверить событие `UserRegistered` → User-service.  

4. **API Gateway:**  
   - Настроить маршрутизацию (Swagger/OpenAPI).  
   - Убедиться, что запросы `/auth/register`, `/users/{id}` проходят через шлюз.  

5. **Recipe и Media:**  
   - Выполнить миграции (Recipe → `recipes`, `recipe_images`, `comments`, `likes`; Media → `media`).  
   - Запустить Media-service, протестировать upload/get/delete (MinIO).  
   - Запустить Recipe-service, проверить создание рецепта с привязкой к media_id, публикацию `RecipeCreated`.  

6. **Chat:**  
   - Выполнить миграции (`conversations`, `conversation_members`, `messages`).  
   - Запустить Chat-service, протестировать создание чата, отправку сообщений (REST/WebSocket).  
   - Проверить публикацию `MessageSent` и реакцию Analytics/Notification.  

7. **Notification:**  
   - Выполнить миграции (`notifications`).  
   - Запустить Notification-service, проверить подписку устройства, получение уведомлений.  
   - Проверить реакцию на события `UserFollowed`, `RecipeCommented`, `MessageSent`.  

8. **Search:**  
   - Развернуть Elasticsearch.  
   - Запустить Search-service, проверить индексацию (прослушивает Kafka) и поиск `/search/recipes?q=...`.  

9. **Analytics:**  
   - Выполнить миграции или настроить ClickHouse-схему.  
   - Запустить Analytics-service, проверить поступление событий из Kafka.  
   - Настроить дашборды в Grafana (подключить к ClickHouse/Prometheus).  

10. **Интеграция и тестирование:**  
    - Проверить все сценарии (регистрация, создание рецепта, комментарии, чат, уведомления, поиск).  
    - Убедиться в отказоустойчивости (останавливать сервис и тестировать поведение).  
    - Провести нагрузочное тестирование (k6, JMeter) и оптимизировать.  

---

**Список используемых технологий (по сервисам):**  
- **Auth:** Gin/Echo, GORM/pgx, PostgreSQL, go-redis, bcrypt, jwt-go, segmentio/kafka-go, zap/logrus, Viper.  
- **User:** Gin/Echo, GORM/pgx, PostgreSQL, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
- **Recipe:** Gin/Echo, GORM/pgx, PostgreSQL, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
- **Chat:** Gin/Echo, Gorilla Websocket/gRPC, PostgreSQL/MongoDB, go-redis/NATS, segmentio/kafka-go, zap/logrus, Viper.  
- **Media:** Gin/Echo, AWS SDK/MinIO, bimg/imaging, FFmpeg, GORM/pgx или MongoDB, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
- **Notification:** Gin/Echo, Firebase Admin SDK, gomail/net/smtp, Twilio, PostgreSQL/MongoDB, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
- **Search:** Gin/Echo, olivere/elastic или meilisearch-go, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
- **Analytics:** Gin/Echo, ClickHouse драйвер или TimescaleDB, prometheus client, segmentio/kafka-go, zap/logrus, Viper.  
- **API Gateway:** Gin/Echo или Kong/Traefik, grpc-gateway, go-redis, segmentio/kafka-go, zap/logrus, Viper.  
