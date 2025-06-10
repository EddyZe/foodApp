--auth service
create schema if not exists auth;

create table if not exists auth.users
(
    id         bigserial primary key,
    email      varchar(255) not null unique,
    password   varchar(255) not null,
    created_at timestamp    not null default now(),
    updated_at  timestamp    not null default now()
);

create table if not exists auth.refresh_token
(
    id         bigserial primary key,
    user_id    bigint references auth.users (id) on delete cascade,
    token      varchar(255) not null unique,
    issue_at   timestamp    not null default now(),
    expired_at timestamp    not null check (expired_at > issue_at),
    is_revoke  bool                  default false
);

create table if not exists auth.users_ban
(
    id         bigserial primary key,
    user_id    bigint references auth.users (id) on delete cascade,
    is_forever bool default false,
    cause      varchar(1024),
    expired_at timestamp not null
);

create table if not exists auth.email_verification_codes
(
    id          bigserial primary key,
    user_id     bigint references auth.users (id) on delete cascade,
    code        varchar(255) not null unique,
    is_verified bool         not null default true,
    create_at   timestamp    not null default now(),
    expired_at  timestamp    not null check (expired_at > create_at)
);

create table if not exists auth.reset_password_codes
(
    id         bigserial primary key,
    user_id    bigint references auth.users (id) on delete cascade,
    code       varchar(255) not null,
    is_valid   bool         not null default true,
    create_at  timestamp    not null default now(),
    expired_at timestamp    not null
);

create table if not exists auth.role
(
    id          bigserial primary key,
    name        varchar(255)  not null unique,
    description varchar(1024) not null
);

create table if not exists auth.permission
(
    id          bigserial primary key,
    name        varchar(255)  not null unique,
    description varchar(1024) not null
);

create table if not exists auth.role_permission
(
    role_id       bigint references auth.role (id) on delete cascade,
    permission_id bigint references auth.permission (id) on delete cascade,
    primary key (role_id, permission_id)
);

create table if not exists auth.user_roles
(
    user_id bigint references auth.users (id) on delete cascade,
    role_id bigint references auth.role (id) on delete cascade,
    primary key (user_id, role_id)
);

create table if not exists auth.black_list_token
(
    token      varchar(255) primary key,
    expired_at timestamp not null
);

create table if not exists auth.audit_log
(
    id         bigserial primary key,
    user_id    bigint       references auth.users (id) on delete set null,
    action     varchar(255) not null,
    ip_address inet,
    user_agent text,
    created_at timestamp    not null default now()
);

create table if not exists auth.password_history
(
    id           bigserial primary key,
    user_id      bigint references auth.users (id) on delete cascade,
    old_password varchar(255) not null,
    changed_at   timestamp    not null default now()
);


create index on auth.users (email);
create index on auth.refresh_token (token);
create index on auth.email_verification_codes (code);
create index on auth.reset_password_codes (code);
create index on auth.users_ban (user_id);
create index on auth.role (name);
create index on auth.permission (name);
create index on auth.audit_log (user_id);
create index on auth.black_list_token (expired_at);
create index on auth.refresh_token (expired_at);
create index on auth.black_list_token (expired_at);
create index on auth.audit_log (created_at);

insert into auth.role(name, description) VALUES ('user', 'Роль обычного пользователя');
insert into auth.role(name, description) VALUES ('admin', 'Роль администратора');

--шедулер
create or replace function auth.clean_expired_data()
    returns void as
$$
begin
    delete from auth.refresh_token where expired_at < NOW();
    delete from auth.black_list_token where expired_at < NOW();
    delete from auth.email_verification_codes where expired_at < NOW();
    delete from auth.reset_password_codes where expired_at < NOW();
    delete from auth.users_ban where expired_at < NOW() and is_forever = false;
end;
$$ language plpgsql;

------------------------------------