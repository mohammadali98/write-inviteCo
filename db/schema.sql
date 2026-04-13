\restrict A9DJXG9WaQpmjK9Mohnf2LANteCrErmH7LN5Aj4TI5pS05m9clfhSnaEVCUEP2k

-- Dumped from database version 16.13 (Homebrew)
-- Dumped by pg_dump version 16.13 (Homebrew)

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
-- Name: card_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.card_images (
    id bigint NOT NULL,
    card_id bigint NOT NULL,
    image text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: card_images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.card_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: card_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.card_images_id_seq OWNED BY public.card_images.id;


--
-- Name: cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cards (
    id bigint NOT NULL,
    name text NOT NULL,
    description text,
    image text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    category text DEFAULT 'wedding-cards'::text NOT NULL,
    price_foil_pkr bigint DEFAULT 0 NOT NULL,
    price_nofoil_pkr bigint DEFAULT 0 NOT NULL,
    price_foil_nok bigint DEFAULT 0 NOT NULL,
    price_nofoil_nok bigint DEFAULT 0 NOT NULL,
    insert_price_pkr bigint DEFAULT 0 NOT NULL,
    insert_price_nok bigint DEFAULT 0 NOT NULL,
    min_order integer DEFAULT 1 NOT NULL,
    included_inserts integer DEFAULT 2 NOT NULL
);


--
-- Name: cards_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cards_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cards_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cards_id_seq OWNED BY public.cards.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id bigint NOT NULL,
    name text NOT NULL,
    email text,
    phone text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    address text,
    city text,
    postal_code text
);


--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: order_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_details (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    bride_name text,
    groom_name text,
    bride_father_name text,
    groom_father_name text,
    event_type text,
    event_date text,
    event_time text,
    rsvp_name text,
    rsvp_phone text,
    notes text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    side text DEFAULT 'bride'::text,
    mehndi_date date,
    baraat_date date,
    nikkah_date date,
    walima_date date,
    mehndi_dinner_time time without time zone,
    baraat_dinner_time time without time zone,
    nikkah_dinner_time time without time zone,
    walima_dinner_time time without time zone,
    mehndi_time_type text,
    mehndi_time time without time zone,
    baraat_time_type text,
    baraat_time time without time zone,
    baraat_arrival_time time without time zone,
    rukhsati_time time without time zone,
    nikkah_time_type text,
    nikkah_time time without time zone,
    walima_time_type text,
    walima_time time without time zone,
    reception_time time without time zone,
    mehndi_day text,
    baraat_day text,
    nikkah_day text,
    walima_day text,
    mehndi_venue_name text,
    mehndi_venue_address text,
    baraat_venue_name text,
    baraat_venue_address text,
    nikkah_venue_name text,
    nikkah_venue_address text,
    walima_venue_name text,
    walima_venue_address text,
    top_label text,
    couple_name text,
    bid_box_details text,
    CONSTRAINT chk_order_details_side CHECK ((side = ANY (ARRAY['bride'::text, 'groom'::text])))
);


--
-- Name: order_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_details_id_seq OWNED BY public.order_details.id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id bigint NOT NULL,
    customer_id bigint,
    card_id bigint,
    quantity bigint DEFAULT 1 NOT NULL,
    total_price bigint NOT NULL,
    status text DEFAULT 'pending'::text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    currency text DEFAULT 'PKR'::text NOT NULL
);


--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: card_images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_images ALTER COLUMN id SET DEFAULT nextval('public.card_images_id_seq'::regclass);


--
-- Name: cards id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cards ALTER COLUMN id SET DEFAULT nextval('public.cards_id_seq'::regclass);


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: order_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_details ALTER COLUMN id SET DEFAULT nextval('public.order_details_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: card_images card_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_images
    ADD CONSTRAINT card_images_pkey PRIMARY KEY (id);


--
-- Name: cards cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cards
    ADD CONSTRAINT cards_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: order_details order_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_details
    ADD CONSTRAINT order_details_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_card_images_card_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_card_images_card_id ON public.card_images USING btree (card_id);


--
-- Name: idx_cards_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cards_category ON public.cards USING btree (category);


--
-- Name: idx_customers_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_email ON public.customers USING btree (email);


--
-- Name: idx_order_details_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_details_order_id ON public.order_details USING btree (order_id);


--
-- Name: idx_orders_card_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_card_id ON public.orders USING btree (card_id);


--
-- Name: idx_orders_customer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_customer_id ON public.orders USING btree (customer_id);


--
-- Name: idx_orders_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_status ON public.orders USING btree (status);


--
-- Name: card_images card_images_card_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_images
    ADD CONSTRAINT card_images_card_id_fkey FOREIGN KEY (card_id) REFERENCES public.cards(id) ON DELETE CASCADE;


--
-- Name: order_details order_details_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_details
    ADD CONSTRAINT order_details_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: orders orders_card_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_card_id_fkey FOREIGN KEY (card_id) REFERENCES public.cards(id);


--
-- Name: orders orders_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- PostgreSQL database dump complete
--

\unrestrict A9DJXG9WaQpmjK9Mohnf2LANteCrErmH7LN5Aj4TI5pS05m9clfhSnaEVCUEP2k


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20260323013540'),
    ('20260323013541'),
    ('20260323013542'),
    ('20260326000001'),
    ('20260326000002'),
    ('20260326000003'),
    ('20260326000004'),
    ('20260326000005'),
    ('20260326000006'),
    ('20260326000007'),
    ('20260326000008'),
    ('20260405111718'),
    ('20260405190000'),
    ('20260406000001'),
    ('20260413000001');
