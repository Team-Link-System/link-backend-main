```
link-backend-main
├─ .dockerignore
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
├─ docker-compose2.yml
├─ go.mod
├─ go.sum
├─ infrastructure
│  ├─ model
│  │  ├─ chat_model.go
│  │  ├─ chatroom_model.go
│  │  ├─ comment_model.go
│  │  ├─ company_model.go
│  │  ├─ department_model.go
│  │  ├─ imogi_model.go
│  │  ├─ like_model.go
│  │  ├─ notification_model.go
│  │  ├─ position_model.go
│  │  ├─ post_model.go
│  │  ├─ postimage_model.go
│  │  ├─ team_model.go
│  │  ├─ user_model.go
│  │  └─ userprofile_model.go
│  └─ persistence
│     ├─ auth_persistence.go
│     ├─ chat_persistence.go
│     ├─ company_persistence.go
│     ├─ depmartment_persistence.go
│     ├─ notification_persistence.go
│     ├─ post_persistence.go
│     ├─ team_persistence.go
│     └─ user_persistence.go
├─ internal
│  ├─ admin
│  │  └─ usecase
│  │     └─ admin_usecase.go
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
│  ├─ company
│  │  ├─ entity
│  │  │  └─ company_entity.go
│  │  ├─ repository
│  │  │  └─ company_repository.go
│  │  └─ usecase
│  │     └─ company_usecase.go
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
│  ├─ post
│  │  ├─ entity
│  │  │  └─ post_entity.go
│  │  ├─ repository
│  │  │  └─ post_repository.go
│  │  └─ usecase
│  │     └─ post_usecase.go
│  ├─ team
│  │  ├─ entity
│  │  │  └─ team_entity.go
│  │  ├─ repository
│  │  │  └─ team_repository.go
│  │  └─ usecase
│  │     └─ team_usecase.go
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
   │  │  ├─ admin_req.go
   │  │  ├─ auth_req.go
   │  │  ├─ chat_req.go
   │  │  ├─ company_req.go
   │  │  ├─ department_req.go
   │  │  ├─ notification_req.go
   │  │  ├─ post_req.go
   │  │  └─ user_req.go
   │  └─ res
   │     ├─ admin_res.go
   │     ├─ auth_res.go
   │     ├─ chat_res.go
   │     ├─ company_res.go
   │     ├─ department_res.go
   │     ├─ notification_res.go
   │     ├─ post_res.go
   │     ├─ team_res.go
   │     ├─ user_res.go
   │     └─ ws_res.go
   ├─ http
   │  ├─ admin_handler.go
   │  ├─ auth_handler.go
   │  ├─ chat_handler.go
   │  ├─ company_handler.go
   │  ├─ department_handler.go
   │  ├─ notification_handler.go
   │  ├─ post_handler.go
   │  ├─ team_handler.go
   │  └─ user_handler.go
   ├─ interceptor
   │  ├─ error_handler.go
   │  └─ token_interceptor.go
   ├─ middleware
   │  └─ image_middleware.go
   ├─ util
   │  ├─ jwt.go
   │  ├─ optional.go
   │  └─ password.go
   └─ ws
      ├─ ws_handler.go
      └─ ws_hub.go

```