# Реализация REST API

* Приложение, хранит данные о пользователе, а также конфиденциальную информацию каждого зарегистрированного участника.  
* Для получения доступа к конфиденциальной информации, а также изменения данных пользователя - необходима авторизация.

* Инструменты:
  1. DB - Postgress 
  2. REST API - github.com/gorilla/mux, cookie, json, context
  3. _test.go - github.com/stretchr/testify, goroutines

### 1. Data Base structure 

```postgresql
create table if not exists public.users
(
    id              bigint generated always as identity
        primary key,
    login           varchar(200) not null
        unique,
    hashed_password varchar(200) not null,
    name            varchar(200) not null,
    surname         varchar(200),
    email           varchar(200) not null
        unique
);

create table if not exists public.info
(
    id     bigint not null
        unique
        references public.users,
    secret text
);
```

### 2. REST API structure

```txt
|_cnd
| |_bellerophon.go  // main function
|
|_iternal
| |_api 
| | |_app.go        // business logic & router binding
| | |_app_test.go
| |  
| |_connect
| | |_connect.go        // soft for connect to DB        
| | |_connectData.json  // data for connect to DB
| | 
| |_source  
|   |_cookie.go
|   |_source.go         // DB operation
|   |_source_test.go
|   |_user.go           // data models define
|
|_ go.mod     
```

### 3. Тесты 

* DB  
1. Создание, получение, удаление пользователя. (users,info)  
2. Изменение: логина, пароля, имени, почты. (users) 
3. Создание, удаление уникальной информации.(info. users)  
 
* REST API  
1. Создание новго пользователя. (SignUp)
2. Логин, авторизация, получение инофрмации (from info) через http.Redirect.
3. Логин, авторизация изменение логина, удаление cookie.

