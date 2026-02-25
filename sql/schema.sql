SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
CREATE SCHEMA public;
ALTER SCHEMA public OWNER TO pg_database_owner;
COMMENT ON SCHEMA public IS 'standard public schema';
SET default_tablespace = '';
SET default_table_access_method = heap;
CREATE TABLE public.daily_schedules (
    user_id uuid NOT NULL,
    date timestamp with time zone NOT NULL,
    max_capacity integer NOT NULL,
    version integer DEFAULT 1 NOT NULL
);
ALTER TABLE public.daily_schedules OWNER TO postgres;
CREATE TABLE public.idempotency_keys (
    id uuid NOT NULL,
    response_status integer NOT NULL,
    response_headers jsonb NOT NULL,
    response_body bytea NOT NULL,
    locked_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.idempotency_keys OWNER TO postgres;
CREATE TABLE public.outbox (
    id uuid NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    last_error text,
    retries integer DEFAULT 0 NOT NULL,
    aggregate_type text DEFAULT ''::text NOT NULL,
    aggregate_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    headers jsonb DEFAULT '{}'::jsonb NOT NULL,
    last_attempted_at timestamp with time zone
);
ALTER TABLE public.outbox OWNER TO postgres;
CREATE TABLE public.schedule_tasks (
    user_id uuid NOT NULL,
    date timestamp with time zone NOT NULL,
    todo_id uuid NOT NULL,
    energy_cost integer NOT NULL
);
ALTER TABLE public.schedule_tasks OWNER TO postgres;
CREATE TABLE public.tags (
    id uuid NOT NULL,
    name text NOT NULL,
    workspace_id uuid NOT NULL
);
ALTER TABLE public.tags OWNER TO postgres;
CREATE TABLE public.todo_completion_logs (
    id uuid NOT NULL,
    todo_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.todo_completion_logs OWNER TO postgres;
CREATE TABLE public.todo_focus_sessions (
    id uuid NOT NULL,
    todo_id uuid NOT NULL,
    user_id uuid,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone
);
ALTER TABLE public.todo_focus_sessions OWNER TO postgres;
CREATE TABLE public.todo_tags (
    todo_id uuid NOT NULL,
    tag_id uuid NOT NULL
);
ALTER TABLE public.todo_tags OWNER TO postgres;
CREATE TABLE public.todos (
    id uuid NOT NULL,
    title text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    workspace_id uuid NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    due_date timestamp with time zone,
    recurrence_interval text,
    recurrence_amount integer,
    last_completed_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.todos OWNER TO postgres;
CREATE TABLE public.user_auth (
    user_id uuid NOT NULL,
    password_hash text,
    totp_status text DEFAULT 'DISABLED'::text NOT NULL,
    totp_secret_cipher bytea,
    totp_secret_nonce bytea
);
ALTER TABLE public.user_auth OWNER TO postgres;
CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.users OWNER TO postgres;
CREATE TABLE public.workspace_members (
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text NOT NULL
);
ALTER TABLE public.workspace_members OWNER TO postgres;
CREATE TABLE public.workspaces (
    id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.workspaces OWNER TO postgres;
ALTER TABLE ONLY public.daily_schedules
    ADD CONSTRAINT daily_schedules_pkey PRIMARY KEY (user_id, date);
ALTER TABLE ONLY public.idempotency_keys
    ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.outbox
    ADD CONSTRAINT outbox_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.schedule_tasks
    ADD CONSTRAINT schedule_tasks_pkey PRIMARY KEY (user_id, date, todo_id);
ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_workspace_id_name_key UNIQUE (workspace_id, name);
ALTER TABLE ONLY public.todo_completion_logs
    ADD CONSTRAINT todo_completion_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.todo_focus_sessions
    ADD CONSTRAINT todo_focus_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.todo_tags
    ADD CONSTRAINT todo_tags_pkey PRIMARY KEY (todo_id, tag_id);
ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.user_auth
    ADD CONSTRAINT user_auth_pkey PRIMARY KEY (user_id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT workspace_members_pkey PRIMARY KEY (workspace_id, user_id);
ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);
CREATE INDEX idx_outbox_unprocessed ON public.outbox USING btree (created_at) WHERE (processed_at IS NULL);
CREATE INDEX idx_todos_workspace_id ON public.todos USING btree (workspace_id);
CREATE INDEX idx_todos_workspace_updated_at ON public.todos USING btree (workspace_id, updated_at);
CREATE UNIQUE INDEX one_active_session_per_todo ON public.todo_focus_sessions USING btree (todo_id) WHERE (end_time IS NULL);
ALTER TABLE ONLY public.daily_schedules
    ADD CONSTRAINT fk_daily_schedules_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.schedule_tasks
    ADD CONSTRAINT fk_schedule_tasks_schedule FOREIGN KEY (user_id, date) REFERENCES public.daily_schedules(user_id, date) ON DELETE CASCADE;
ALTER TABLE ONLY public.tags
    ADD CONSTRAINT fk_tags_workspace_id FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todo_completion_logs
    ADD CONSTRAINT fk_todo_completion_logs_actor_id FOREIGN KEY (actor_id) REFERENCES public.users(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.todo_completion_logs
    ADD CONSTRAINT fk_todo_completion_logs_todo_id FOREIGN KEY (todo_id) REFERENCES public.todos(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todo_focus_sessions
    ADD CONSTRAINT fk_todo_focus_sessions_todo_id FOREIGN KEY (todo_id) REFERENCES public.todos(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todo_focus_sessions
    ADD CONSTRAINT fk_todo_focus_sessions_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todo_tags
    ADD CONSTRAINT fk_todo_tags_tag_id FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todo_tags
    ADD CONSTRAINT fk_todo_tags_todo_id FOREIGN KEY (todo_id) REFERENCES public.todos(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.todos
    ADD CONSTRAINT fk_todos_workspace_id FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.user_auth
    ADD CONSTRAINT fk_user_auth_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_wm_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_wm_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id) ON DELETE CASCADE;
