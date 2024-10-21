
```
link
├─ .dockerignore
├─ .gitignore
├─ Dockerfile
├─ EnvKey
├─ README.md
├─ cmd
│  └─ main.go
├─ config
│  ├─ config.go
│  ├─ di.go
│  └─ init.go
├─ docker-compose.yml
├─ go.mod
├─ go.sum
├─ infrastructure
│  ├─ logger
│  ├─ model
│  │  ├─ chat_model.go
│  │  ├─ chatroom_model.go
│  │  ├─ comment_model.go
│  │  ├─ company_model.go
│  │  ├─ department_model.go
│  │  ├─ like_model.go
│  │  ├─ notification_model.go
│  │  ├─ position_model.go
│  │  ├─ post_model.go
│  │  ├─ team_model.go
│  │  ├─ user_model.go
│  │  └─ userprofile_model.go
│  └─ persistence
│     ├─ auth_persistence_redis.go
│     ├─ chat_persistence.go
│     ├─ depmartment_persistence_pg.go
│     ├─ notification_persistence_mongo.go
│     └─ user_persistence_pg.go
├─ internal
│  ├─ auth
│  │  ├─ entity
│  │  │  └─ token_entity.go
│  │  ├─ repository
│  │  │  └─ auth_repository.go
│  │  └─ usecase
│  │     └─ auth_usecase.go
│  ├─ chat
│  │  ├─ entity
│  │  │  └─ chat_entity.go
│  │  ├─ repository
│  │  │  └─ chat_repository.go
│  │  └─ usecase
│  │     └─ chat_usecase.go
│  ├─ department
│  │  ├─ entity
│  │  │  └─ department.go
│  │  ├─ repository
│  │  │  └─ department_repository.go
│  │  └─ usecase
│  │     └─ department_usecase.go
│  ├─ notification
│  │  ├─ entity
│  │  │  └─ notification_entity.go
│  │  ├─ repository
│  │  │  └─ notification_repository.go
│  │  └─ usecase
│  │     └─ notification_usecase.go
│  ├─ team
│  │  ├─ entity
│  │  ├─ repository
│  │  └─ usecase
│  └─ user
│     ├─ entity
│     │  └─ user_entity.go
│     ├─ repository
│     │  └─ user_repository.go
│     └─ usecase
│        └─ user_usecase.go
└─ pkg
   ├─ common
   │  └─ response.go
   ├─ dto
   │  ├─ req
   │  │  ├─ auth_req.go
   │  │  ├─ chat_req.go
   │  │  ├─ department_req.go
   │  │  ├─ notification_req.go
   │  │  └─ user_req.go
   │  └─ res
   │     ├─ auth_res.go
   │     ├─ chat_res.go
   │     ├─ department_res.go
   │     ├─ notification_res.go
   │     ├─ user_res.go
   │     └─ ws_res.go
   ├─ http
   │  ├─ auth_handler.go
   │  ├─ chat_handler.go
   │  ├─ department_handler.go
   │  ├─ notification_handler.go
   │  └─ user_handler.go
   ├─ interceptor
   │  ├─ error_handler.go
   │  └─ token_interceptor.go
   ├─ util
   │  ├─ jwt.go
   │  └─ password.go
   └─ ws
      ├─ ws_handler.go
      └─ ws_hub.go

```