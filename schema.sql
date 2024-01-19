--
-- PostgreSQL database dump
--

-- Dumped from database version 16.0 (Debian 16.0-1.pgdg120+1)
-- Dumped by pg_dump version 16.0 (Debian 16.0-1.pgdg120+1)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.migrations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(255) NOT NULL,
    source_id bigint NOT NULL,
    destination_id bigint NOT NULL,
    is_success boolean NOT NULL,
    last_run_at timestamp with time zone NOT NULL,
    is_proxy_mapped boolean NOT NULL,
    has_template_bindings boolean NOT NULL,
    is_proxy_imported boolean DEFAULT false NOT NULL,
    default_proxy_id bigint,
    is_template_imported boolean DEFAULT false NOT NULL,
    is_running boolean DEFAULT false NOT NULL,
    is_template_successful boolean DEFAULT false NOT NULL,
    is_template_running boolean DEFAULT false NOT NULL,
    is_default_running boolean DEFAULT false NOT NULL,
    is_default_successful boolean DEFAULT false NOT NULL,
    is_default_host_importing boolean DEFAULT false NOT NULL,
    is_default_host_imported boolean DEFAULT false NOT NULL,
    is_default_disabling boolean DEFAULT false NOT NULL,
    is_default_disabled boolean DEFAULT false NOT NULL,
    is_default_rolling_back boolean DEFAULT false NOT NULL
);


ALTER TABLE public.migrations OWNER TO crodont;

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.migrations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.migrations_id_seq OWNER TO crodont;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: zabbix_hosts; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_hosts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    host_id character varying(255) NOT NULL,
    host character varying(255) NOT NULL,
    proxy_host_id character varying(255),
    status text NOT NULL,
    migration_id bigint NOT NULL,
    disabled bigint NOT NULL
);


ALTER TABLE public.zabbix_hosts OWNER TO crodont;

--
-- Name: zabbix_hosts_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_hosts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_hosts_id_seq OWNER TO crodont;

--
-- Name: zabbix_hosts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_hosts_id_seq OWNED BY public.zabbix_hosts.id;


--
-- Name: zabbix_parent_templates; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_parent_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    template_id bigint NOT NULL,
    host text NOT NULL,
    child_id bigint NOT NULL
);


ALTER TABLE public.zabbix_parent_templates OWNER TO crodont;

--
-- Name: zabbix_parent_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_parent_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_parent_templates_id_seq OWNER TO crodont;

--
-- Name: zabbix_parent_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_parent_templates_id_seq OWNED BY public.zabbix_parent_templates.id;


--
-- Name: zabbix_proxies; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_proxies (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    host character varying(255) NOT NULL,
    status text NOT NULL,
    last_access text NOT NULL,
    proxy_address character varying(255),
    host_count bigint NOT NULL,
    interface_id bigint,
    migration_id bigint NOT NULL,
    zabbix_server_id bigint NOT NULL,
    proxy_mapping_id bigint,
    proxy_id character varying(255),
    is_hosts_running boolean DEFAULT false NOT NULL,
    is_host_successful boolean DEFAULT false NOT NULL,
    is_host_importing boolean DEFAULT false NOT NULL,
    is_host_imported boolean DEFAULT false NOT NULL,
    is_host_disabling boolean DEFAULT false NOT NULL,
    is_rolling_back boolean DEFAULT false NOT NULL,
    is_host_disabled boolean DEFAULT false NOT NULL
);


ALTER TABLE public.zabbix_proxies OWNER TO crodont;

--
-- Name: zabbix_proxies_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_proxies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_proxies_id_seq OWNER TO crodont;

--
-- Name: zabbix_proxies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_proxies_id_seq OWNED BY public.zabbix_proxies.id;


--
-- Name: zabbix_proxy_interfaces; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_proxy_interfaces (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dns character varying(255) NOT NULL,
    ip character varying(255) NOT NULL,
    port bigint NOT NULL,
    interfaceid character varying(255) NOT NULL
);


ALTER TABLE public.zabbix_proxy_interfaces OWNER TO crodont;

--
-- Name: zabbix_proxy_interfaces_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_proxy_interfaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_proxy_interfaces_id_seq OWNER TO crodont;

--
-- Name: zabbix_proxy_interfaces_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_proxy_interfaces_id_seq OWNED BY public.zabbix_proxy_interfaces.id;


--
-- Name: zabbix_proxy_mappings; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_proxy_mappings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    source_proxy_id bigint NOT NULL,
    destination_proxy_id bigint NOT NULL
);


ALTER TABLE public.zabbix_proxy_mappings OWNER TO crodont;

--
-- Name: zabbix_proxy_mappings_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_proxy_mappings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_proxy_mappings_id_seq OWNER TO crodont;

--
-- Name: zabbix_proxy_mappings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_proxy_mappings_id_seq OWNED BY public.zabbix_proxy_mappings.id;


--
-- Name: zabbix_servers; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_servers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    url text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    version bigint NOT NULL
);


ALTER TABLE public.zabbix_servers OWNER TO crodont;

--
-- Name: zabbix_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_servers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_servers_id_seq OWNER TO crodont;

--
-- Name: zabbix_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_servers_id_seq OWNED BY public.zabbix_servers.id;


--
-- Name: zabbix_template_mappings; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_template_mappings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    source_template_id bigint NOT NULL,
    destination_template_id bigint NOT NULL,
    is_new boolean DEFAULT false
);


ALTER TABLE public.zabbix_template_mappings OWNER TO crodont;

--
-- Name: zabbix_template_mappings_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_template_mappings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_template_mappings_id_seq OWNER TO crodont;

--
-- Name: zabbix_template_mappings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_template_mappings_id_seq OWNED BY public.zabbix_template_mappings.id;


--
-- Name: zabbix_templates; Type: TABLE; Schema: public; Owner: crodont
--

CREATE TABLE public.zabbix_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    templateid character varying(255) NOT NULL,
    name text NOT NULL,
    host text NOT NULL,
    description text NOT NULL,
    host_count bigint NOT NULL,
    items bigint NOT NULL,
    triggers bigint NOT NULL,
    graphs bigint NOT NULL,
    screens bigint NOT NULL,
    discoveries bigint NOT NULL,
    http_tests bigint NOT NULL,
    macros bigint NOT NULL,
    migration_id bigint NOT NULL,
    zabbix_server_id bigint NOT NULL,
    remote_found text
);


ALTER TABLE public.zabbix_templates OWNER TO crodont;

--
-- Name: zabbix_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: crodont
--

CREATE SEQUENCE public.zabbix_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.zabbix_templates_id_seq OWNER TO crodont;

--
-- Name: zabbix_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: crodont
--

ALTER SEQUENCE public.zabbix_templates_id_seq OWNED BY public.zabbix_templates.id;


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: zabbix_hosts id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_hosts ALTER COLUMN id SET DEFAULT nextval('public.zabbix_hosts_id_seq'::regclass);


--
-- Name: zabbix_parent_templates id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_parent_templates ALTER COLUMN id SET DEFAULT nextval('public.zabbix_parent_templates_id_seq'::regclass);


--
-- Name: zabbix_proxies id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies ALTER COLUMN id SET DEFAULT nextval('public.zabbix_proxies_id_seq'::regclass);


--
-- Name: zabbix_proxy_interfaces id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_interfaces ALTER COLUMN id SET DEFAULT nextval('public.zabbix_proxy_interfaces_id_seq'::regclass);


--
-- Name: zabbix_proxy_mappings id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings ALTER COLUMN id SET DEFAULT nextval('public.zabbix_proxy_mappings_id_seq'::regclass);


--
-- Name: zabbix_servers id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_servers ALTER COLUMN id SET DEFAULT nextval('public.zabbix_servers_id_seq'::regclass);


--
-- Name: zabbix_template_mappings id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings ALTER COLUMN id SET DEFAULT nextval('public.zabbix_template_mappings_id_seq'::regclass);


--
-- Name: zabbix_templates id; Type: DEFAULT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_templates ALTER COLUMN id SET DEFAULT nextval('public.zabbix_templates_id_seq'::regclass);


--
-- Name: zabbix_servers idx_zabbix_servers_name; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_servers
    ADD CONSTRAINT idx_zabbix_servers_name UNIQUE (name);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: zabbix_hosts zabbix_hosts_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_hosts
    ADD CONSTRAINT zabbix_hosts_pkey PRIMARY KEY (id);


--
-- Name: zabbix_parent_templates zabbix_parent_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_parent_templates
    ADD CONSTRAINT zabbix_parent_templates_pkey PRIMARY KEY (id);


--
-- Name: zabbix_proxies zabbix_proxies_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies
    ADD CONSTRAINT zabbix_proxies_pkey PRIMARY KEY (id);


--
-- Name: zabbix_proxy_interfaces zabbix_proxy_interfaces_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_interfaces
    ADD CONSTRAINT zabbix_proxy_interfaces_pkey PRIMARY KEY (id);


--
-- Name: zabbix_proxy_mappings zabbix_proxy_mappings_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings
    ADD CONSTRAINT zabbix_proxy_mappings_pkey PRIMARY KEY (id);


--
-- Name: zabbix_servers zabbix_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_servers
    ADD CONSTRAINT zabbix_servers_pkey PRIMARY KEY (id);


--
-- Name: zabbix_servers zabbix_servers_url_key; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_servers
    ADD CONSTRAINT zabbix_servers_url_key UNIQUE (url);


--
-- Name: zabbix_template_mappings zabbix_template_mappings_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings
    ADD CONSTRAINT zabbix_template_mappings_pkey PRIMARY KEY (id);


--
-- Name: zabbix_templates zabbix_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_templates
    ADD CONSTRAINT zabbix_templates_pkey PRIMARY KEY (id);


--
-- Name: idx_migrations_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_migrations_deleted_at ON public.migrations USING btree (deleted_at);


--
-- Name: idx_zabbix_hosts_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_hosts_deleted_at ON public.zabbix_hosts USING btree (deleted_at);


--
-- Name: idx_zabbix_hosts_proxy_host_id; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_hosts_proxy_host_id ON public.zabbix_hosts USING btree (proxy_host_id);


--
-- Name: idx_zabbix_parent_templates_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_parent_templates_deleted_at ON public.zabbix_parent_templates USING btree (deleted_at);


--
-- Name: idx_zabbix_proxies_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_proxies_deleted_at ON public.zabbix_proxies USING btree (deleted_at);


--
-- Name: idx_zabbix_proxies_proxy_id; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_proxies_proxy_id ON public.zabbix_proxies USING btree (proxy_id);


--
-- Name: idx_zabbix_proxy_interfaces_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_proxy_interfaces_deleted_at ON public.zabbix_proxy_interfaces USING btree (deleted_at);


--
-- Name: idx_zabbix_proxy_mappings_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_proxy_mappings_deleted_at ON public.zabbix_proxy_mappings USING btree (deleted_at);


--
-- Name: idx_zabbix_servers_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_servers_deleted_at ON public.zabbix_servers USING btree (deleted_at);


--
-- Name: idx_zabbix_template_mappings_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_template_mappings_deleted_at ON public.zabbix_template_mappings USING btree (deleted_at);


--
-- Name: idx_zabbix_templates_deleted_at; Type: INDEX; Schema: public; Owner: crodont
--

CREATE INDEX idx_zabbix_templates_deleted_at ON public.zabbix_templates USING btree (deleted_at);


--
-- Name: migrations fk_migrations_default_proxy; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT fk_migrations_default_proxy FOREIGN KEY (default_proxy_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: migrations fk_migrations_destination; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT fk_migrations_destination FOREIGN KEY (destination_id) REFERENCES public.zabbix_servers(id);


--
-- Name: migrations fk_migrations_source; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT fk_migrations_source FOREIGN KEY (source_id) REFERENCES public.zabbix_servers(id);


--
-- Name: zabbix_hosts fk_zabbix_hosts_migration; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_hosts
    ADD CONSTRAINT fk_zabbix_hosts_migration FOREIGN KEY (migration_id) REFERENCES public.migrations(id);


--
-- Name: zabbix_proxy_mappings fk_zabbix_proxies_destination_mapping; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings
    ADD CONSTRAINT fk_zabbix_proxies_destination_mapping FOREIGN KEY (destination_proxy_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: zabbix_proxies fk_zabbix_proxies_interface; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies
    ADD CONSTRAINT fk_zabbix_proxies_interface FOREIGN KEY (interface_id) REFERENCES public.zabbix_proxy_interfaces(id);


--
-- Name: zabbix_proxies fk_zabbix_proxies_migration; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies
    ADD CONSTRAINT fk_zabbix_proxies_migration FOREIGN KEY (migration_id) REFERENCES public.migrations(id);


--
-- Name: zabbix_proxies fk_zabbix_proxies_proxy_mapping; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies
    ADD CONSTRAINT fk_zabbix_proxies_proxy_mapping FOREIGN KEY (proxy_mapping_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: zabbix_proxy_mappings fk_zabbix_proxies_source_mapping; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings
    ADD CONSTRAINT fk_zabbix_proxies_source_mapping FOREIGN KEY (source_proxy_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: zabbix_proxies fk_zabbix_proxies_zabbix_server; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxies
    ADD CONSTRAINT fk_zabbix_proxies_zabbix_server FOREIGN KEY (zabbix_server_id) REFERENCES public.zabbix_servers(id);


--
-- Name: zabbix_proxy_mappings fk_zabbix_proxy_mappings_destination_proxy; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings
    ADD CONSTRAINT fk_zabbix_proxy_mappings_destination_proxy FOREIGN KEY (destination_proxy_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: zabbix_proxy_mappings fk_zabbix_proxy_mappings_source_proxy; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_proxy_mappings
    ADD CONSTRAINT fk_zabbix_proxy_mappings_source_proxy FOREIGN KEY (source_proxy_id) REFERENCES public.zabbix_proxies(id);


--
-- Name: zabbix_template_mappings fk_zabbix_template_mappings_destination_template; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings
    ADD CONSTRAINT fk_zabbix_template_mappings_destination_template FOREIGN KEY (destination_template_id) REFERENCES public.zabbix_templates(id);


--
-- Name: zabbix_template_mappings fk_zabbix_template_mappings_source_template; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings
    ADD CONSTRAINT fk_zabbix_template_mappings_source_template FOREIGN KEY (source_template_id) REFERENCES public.zabbix_templates(id);


--
-- Name: zabbix_template_mappings fk_zabbix_templates_destination_mapping; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings
    ADD CONSTRAINT fk_zabbix_templates_destination_mapping FOREIGN KEY (destination_template_id) REFERENCES public.zabbix_templates(id);


--
-- Name: zabbix_templates fk_zabbix_templates_migration; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_templates
    ADD CONSTRAINT fk_zabbix_templates_migration FOREIGN KEY (migration_id) REFERENCES public.migrations(id);


--
-- Name: zabbix_parent_templates fk_zabbix_templates_parents; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_parent_templates
    ADD CONSTRAINT fk_zabbix_templates_parents FOREIGN KEY (child_id) REFERENCES public.zabbix_templates(id);


--
-- Name: zabbix_template_mappings fk_zabbix_templates_source_mapping; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_template_mappings
    ADD CONSTRAINT fk_zabbix_templates_source_mapping FOREIGN KEY (source_template_id) REFERENCES public.zabbix_templates(id);


--
-- Name: zabbix_templates fk_zabbix_templates_zabbix_server; Type: FK CONSTRAINT; Schema: public; Owner: crodont
--

ALTER TABLE ONLY public.zabbix_templates
    ADD CONSTRAINT fk_zabbix_templates_zabbix_server FOREIGN KEY (zabbix_server_id) REFERENCES public.zabbix_servers(id);


--
-- PostgreSQL database dump complete
--

