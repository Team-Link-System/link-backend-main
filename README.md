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
│  │  ├─ department_model.go
│  │  ├─ group_model.go
│  │  └─ user_model.go
│  └─ persistence
│     ├─ auth_persistence_redis.go
│     ├─ depmartment_persistence_pg.go
│     └─ user_persistence_pg.go
├─ internal
│  ├─ auth
│  │  ├─ entity
│  │  │  └─ token_entity.go
│  │  ├─ repository
│  │  │  └─ auth_repository.go
│  │  └─ usecase
│  │     └─ auth_usecase.go
│  ├─ department
│  │  ├─ entity
│  │  │  └─ department.go
│  │  ├─ repository
│  │  │  └─ department_repository.go
│  │  └─ usecase
│  │     └─ department_usecase.go
│  ├─ group
│  │  ├─ entity
│  │  │  └─ group_entity.go
│  │  ├─ repository
│  │  └─ usecase
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
   ├─ dto
   │  ├─ auth
   │  │  ├─ req
   │  │  │  └─ auth_req.go
   │  │  └─ res
   │  │     └─ auth_res.go
   │  ├─ department
   │  │  ├─ req
   │  │  │  └─ department_req.go
   │  │  └─ res
   │  │     └─ department_res.go
   │  └─ user
   │     ├─ req
   │     │  └─ user_req.go
   │     └─ res
   │        └─ user_res.go
   ├─ http
   │  ├─ auth_handler.go
   │  ├─ department_handler.go
   │  └─ user_handler.go
   ├─ interceptor
   │  ├─ error_handler.go
   │  ├─ response.go
   │  └─ token_interceptor.go
   ├─ util
   │  ├─ jwt.go
   │  └─ password.go
   └─ ws
      └─ websocket.go

```