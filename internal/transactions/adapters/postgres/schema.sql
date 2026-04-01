\restrict dbmate

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id uuid NOT NULL,
    user_id character varying(255) NOT NULL,
    amount bigint NOT NULL,
    transaction_type character varying(10) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
)
PARTITION BY RANGE (id);


--
-- Name: transactions_default; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions_default (
    id uuid CONSTRAINT transactions_id_not_null NOT NULL,
    user_id character varying(255) CONSTRAINT transactions_user_id_not_null NOT NULL,
    amount bigint CONSTRAINT transactions_amount_not_null NOT NULL,
    transaction_type character varying(10) CONSTRAINT transactions_transaction_type_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT transactions_created_at_not_null NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
);


--
-- Name: transactions_y2026_m04; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions_y2026_m04 (
    id uuid CONSTRAINT transactions_id_not_null NOT NULL,
    user_id character varying(255) CONSTRAINT transactions_user_id_not_null NOT NULL,
    amount bigint CONSTRAINT transactions_amount_not_null NOT NULL,
    transaction_type character varying(10) CONSTRAINT transactions_transaction_type_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT transactions_created_at_not_null NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
);


--
-- Name: transactions_y2026_m05; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions_y2026_m05 (
    id uuid CONSTRAINT transactions_id_not_null NOT NULL,
    user_id character varying(255) CONSTRAINT transactions_user_id_not_null NOT NULL,
    amount bigint CONSTRAINT transactions_amount_not_null NOT NULL,
    transaction_type character varying(10) CONSTRAINT transactions_transaction_type_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT transactions_created_at_not_null NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
);


--
-- Name: transactions_y2026_m06; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions_y2026_m06 (
    id uuid CONSTRAINT transactions_id_not_null NOT NULL,
    user_id character varying(255) CONSTRAINT transactions_user_id_not_null NOT NULL,
    amount bigint CONSTRAINT transactions_amount_not_null NOT NULL,
    transaction_type character varying(10) CONSTRAINT transactions_transaction_type_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT transactions_created_at_not_null NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
);


--
-- Name: transactions_y2026_m07; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions_y2026_m07 (
    id uuid CONSTRAINT transactions_id_not_null NOT NULL,
    user_id character varying(255) CONSTRAINT transactions_user_id_not_null NOT NULL,
    amount bigint CONSTRAINT transactions_amount_not_null NOT NULL,
    transaction_type character varying(10) CONSTRAINT transactions_transaction_type_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT transactions_created_at_not_null NOT NULL,
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY ((ARRAY['bet'::character varying, 'win'::character varying])::text[])))
);


--
-- Name: transactions_default; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ATTACH PARTITION public.transactions_default DEFAULT;


--
-- Name: transactions_y2026_m04; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ATTACH PARTITION public.transactions_y2026_m04 FOR VALUES FROM ('019d4657-0000-7000-8000-000000000000') TO ('019de0d5-c800-7000-8000-000000000000');


--
-- Name: transactions_y2026_m05; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ATTACH PARTITION public.transactions_y2026_m05 FOR VALUES FROM ('019de0d5-c800-7000-8000-000000000000') TO ('019e807a-ec00-7000-8000-000000000000');


--
-- Name: transactions_y2026_m06; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ATTACH PARTITION public.transactions_y2026_m06 FOR VALUES FROM ('019e807a-ec00-7000-8000-000000000000') TO ('019f1af9-b400-7000-8000-000000000000');


--
-- Name: transactions_y2026_m07; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ATTACH PARTITION public.transactions_y2026_m07 FOR VALUES FROM ('019f1af9-b400-7000-8000-000000000000') TO ('019fba9e-d800-7000-8000-000000000000');


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: transactions_default transactions_default_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions_default
    ADD CONSTRAINT transactions_default_pkey PRIMARY KEY (id);


--
-- Name: transactions_y2026_m04 transactions_y2026_m04_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions_y2026_m04
    ADD CONSTRAINT transactions_y2026_m04_pkey PRIMARY KEY (id);


--
-- Name: transactions_y2026_m05 transactions_y2026_m05_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions_y2026_m05
    ADD CONSTRAINT transactions_y2026_m05_pkey PRIMARY KEY (id);


--
-- Name: transactions_y2026_m06 transactions_y2026_m06_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions_y2026_m06
    ADD CONSTRAINT transactions_y2026_m06_pkey PRIMARY KEY (id);


--
-- Name: transactions_y2026_m07 transactions_y2026_m07_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions_y2026_m07
    ADD CONSTRAINT transactions_y2026_m07_pkey PRIMARY KEY (id);


--
-- Name: idx_transactions_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_type_id ON ONLY public.transactions USING btree (transaction_type, id DESC);


--
-- Name: idx_transactions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_user_id ON ONLY public.transactions USING btree (user_id, id DESC);


--
-- Name: transactions_default_transaction_type_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_default_transaction_type_id_idx ON public.transactions_default USING btree (transaction_type, id DESC);


--
-- Name: transactions_default_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_default_user_id_id_idx ON public.transactions_default USING btree (user_id, id DESC);


--
-- Name: transactions_y2026_m04_transaction_type_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m04_transaction_type_id_idx ON public.transactions_y2026_m04 USING btree (transaction_type, id DESC);


--
-- Name: transactions_y2026_m04_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m04_user_id_id_idx ON public.transactions_y2026_m04 USING btree (user_id, id DESC);


--
-- Name: transactions_y2026_m05_transaction_type_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m05_transaction_type_id_idx ON public.transactions_y2026_m05 USING btree (transaction_type, id DESC);


--
-- Name: transactions_y2026_m05_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m05_user_id_id_idx ON public.transactions_y2026_m05 USING btree (user_id, id DESC);


--
-- Name: transactions_y2026_m06_transaction_type_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m06_transaction_type_id_idx ON public.transactions_y2026_m06 USING btree (transaction_type, id DESC);


--
-- Name: transactions_y2026_m06_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m06_user_id_id_idx ON public.transactions_y2026_m06 USING btree (user_id, id DESC);


--
-- Name: transactions_y2026_m07_transaction_type_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m07_transaction_type_id_idx ON public.transactions_y2026_m07 USING btree (transaction_type, id DESC);


--
-- Name: transactions_y2026_m07_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_y2026_m07_user_id_id_idx ON public.transactions_y2026_m07 USING btree (user_id, id DESC);


--
-- Name: transactions_default_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.transactions_pkey ATTACH PARTITION public.transactions_default_pkey;


--
-- Name: transactions_default_transaction_type_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_type_id ATTACH PARTITION public.transactions_default_transaction_type_id_idx;


--
-- Name: transactions_default_user_id_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_user_id ATTACH PARTITION public.transactions_default_user_id_id_idx;


--
-- Name: transactions_y2026_m04_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.transactions_pkey ATTACH PARTITION public.transactions_y2026_m04_pkey;


--
-- Name: transactions_y2026_m04_transaction_type_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_type_id ATTACH PARTITION public.transactions_y2026_m04_transaction_type_id_idx;


--
-- Name: transactions_y2026_m04_user_id_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_user_id ATTACH PARTITION public.transactions_y2026_m04_user_id_id_idx;


--
-- Name: transactions_y2026_m05_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.transactions_pkey ATTACH PARTITION public.transactions_y2026_m05_pkey;


--
-- Name: transactions_y2026_m05_transaction_type_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_type_id ATTACH PARTITION public.transactions_y2026_m05_transaction_type_id_idx;


--
-- Name: transactions_y2026_m05_user_id_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_user_id ATTACH PARTITION public.transactions_y2026_m05_user_id_id_idx;


--
-- Name: transactions_y2026_m06_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.transactions_pkey ATTACH PARTITION public.transactions_y2026_m06_pkey;


--
-- Name: transactions_y2026_m06_transaction_type_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_type_id ATTACH PARTITION public.transactions_y2026_m06_transaction_type_id_idx;


--
-- Name: transactions_y2026_m06_user_id_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_user_id ATTACH PARTITION public.transactions_y2026_m06_user_id_id_idx;


--
-- Name: transactions_y2026_m07_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.transactions_pkey ATTACH PARTITION public.transactions_y2026_m07_pkey;


--
-- Name: transactions_y2026_m07_transaction_type_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_type_id ATTACH PARTITION public.transactions_y2026_m07_transaction_type_id_idx;


--
-- Name: transactions_y2026_m07_user_id_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_transactions_user_id ATTACH PARTITION public.transactions_y2026_m07_user_id_id_idx;


--
-- PostgreSQL database dump complete
--

\unrestrict dbmate


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('001');
