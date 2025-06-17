-- Удаление функции (если есть)
drop function if exists auth.clean_expired_data() cascade;

-- Удаление таблиц (в обратном порядке зависимости)
drop table if exists auth.password_history cascade;
drop table if exists auth.audit_log cascade;
drop table if exists auth.black_list_token cascade;
drop table if exists auth.user_roles cascade;
drop table if exists auth.role_permission cascade;
drop table if exists auth.permission cascade;
drop table if exists auth.role cascade;
drop table if exists auth.reset_password_codes cascade;
drop table if exists auth.email_verification_codes cascade;
drop table if exists auth.users_ban cascade;
drop table if exists auth.refresh_token cascade ;
drop table if exists auth.refresh_token cascade;
drop table if exists auth.users cascade;

-- Удаление схемы, если нужно полностью очистить
drop schema if exists auth cascade;
