Goods Service

Запуск
* docker-compose up --build 
* make migrate-up-pg #миграции postgresql (лучше накатывать миграции через powershell)
* make migrate-up-ch #миграции clickhouse (лучше накатывать миграции через git bash)

REST API

POST /good/create - создать good/проект/товар
Query ?project_id=1

PATCH /good/update/:id - обновить good/проект/товар
Query ?project_id=1

DELETE /good/remove/:id - мягкое удаление good/проект/товар
Query ?project_id=1

GET /goods/list - список good/проект/товар
Query ?project_id=1&limit=10&offset=0&sort=desc

PATCH /goods/:id/reprioritize - изменение приоритета
Query ?project_id=1

Возможные команды Makefile
* make up               # docker-compose up -d
* make down             # docker-compose down
* make migrate-up-pg    # Миграции PostgreSQL up
* make migrate-down-pg  # Ролбэк PostgreSQL
* make migrate-up-ch    # ClickHouse up
* make migrate-down-ch  # ClickHouse down
* make status-pg        # Список таблиц PostgreSQL
* make status-ch        # Список таблиц ClickHouse

Переменные .env
* POSTGRES_DB=dbname
* POSTGRES_USER=username
* POSTGRES_PASSWORD=pgpassword
* POSTGRES_HOST=pghost
* POSTGRES_PORT=5432

CURL запросы

* curl -X POST "http://localhost:8080/good/create?project_id=1" -H "Content-Type: application/json" -d '{"name":"test_good"} - создать
* curl -X PATCH "http://localhost:8080/good/update/2?project_id=1" -H "Content-Type: application/json" -d '{"name":"patch_test","description":"desc"}' - обновить
* curl -X DELETE "http://localhost:8080/good/remove/2?project_id=1" - удалить (soft delete)
* curl "http://localhost:8080/goods/list?project_id=1&limit=10&offset=0&sort=desc" - получить весь список по project_id
* curl -X PATCH "http://localhost:8080/goods/3/reprioritize?project_id=2" -H "Content-Type: application/json" -d '{"newPriority": 1}' - перераспределение приоритета
