--
-- PostgreSQL database dump
--

-- Dumped from database version 14.13 (Homebrew)
-- Dumped by pg_dump version 14.13 (Homebrew)

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

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: category_enum; Type: TYPE; Schema: public; Owner: killian
--

CREATE TYPE public.category_enum AS ENUM (
    'tops',
    'bottoms',
    'outerwear',
    'footwear',
    'accessories'
);


ALTER TYPE public.category_enum OWNER TO killian;

--
-- Name: item_status_enum; Type: TYPE; Schema: public; Owner: killian
--

CREATE TYPE public.item_status_enum AS ENUM (
    'available',
    'sold',
    'reserved'
);


ALTER TYPE public.item_status_enum OWNER TO killian;

--
-- Name: order_status_enum; Type: TYPE; Schema: public; Owner: killian
--

CREATE TYPE public.order_status_enum AS ENUM (
    'pending',
    'processing',
    'shipped',
    'delivered',
    'cancelled'
);


ALTER TYPE public.order_status_enum OWNER TO killian;

--
-- Name: size_enum; Type: TYPE; Schema: public; Owner: killian
--

CREATE TYPE public.size_enum AS ENUM (
    'XS',
    'S',
    'M',
    'L',
    'XL'
);


ALTER TYPE public.size_enum OWNER TO killian;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.cart_items (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    item_id uuid,
    added_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.cart_items OWNER TO killian;

--
-- Name: item_images; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.item_images (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    item_id uuid,
    image_path character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.item_images OWNER TO killian;

--
-- Name: items; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.items (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    size public.size_enum NOT NULL,
    category public.category_enum NOT NULL,
    status public.item_status_enum DEFAULT 'available'::public.item_status_enum,
    quantity integer DEFAULT 1,
    seller_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT quantity_non_negative CHECK ((quantity >= 0))
);


ALTER TABLE public.items OWNER TO killian;

--
-- Name: order_items; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.order_items (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    order_id uuid,
    item_id uuid,
    price_at_time numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.order_items OWNER TO killian;

--
-- Name: orders; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.orders (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    total numeric(10,2) NOT NULL,
    status public.order_status_enum DEFAULT 'pending'::public.order_status_enum,
    address text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.orders OWNER TO killian;

--
-- Name: users; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.users OWNER TO killian;

--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.cart_items (id, user_id, item_id, added_at) FROM stdin;
acd91127-ff08-4377-ac42-ae7711d74df2	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2ab78fc3-29c7-4f07-9b0d-14678397b4a6	2024-11-27 12:12:54.42967+01
\.


--
-- Data for Name: item_images; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.item_images (id, item_id, image_path, created_at) FROM stdin;
9b74c4d5-ca7b-4213-9063-03ee74603f89	55020891-ded6-4262-bebc-36f5c7f7bbd6	uploads/f232bb94-11a7-4a73-9fe1-8c06cc9a62cb.webp	2024-11-26 17:20:24.222401+01
d035a4e4-f6fb-4ca9-a9a6-3a33378f136a	251a1dd9-3506-4fe6-9416-cd281c89de77	uploads/63f3d685-ee51-47f8-ab31-9b679c5929b3.webp	2024-11-26 17:27:30.152+01
f0a5f2b2-2c62-4877-a27c-32d4eb9d6666	177afae1-a840-4f11-996e-e9f3c8378694	uploads/c75f681a-62dd-4e13-b264-a4a31522b469.webp	2024-11-26 17:44:04.063173+01
a4174939-caa0-44a0-93b7-296b4fd9999a	7e52385e-4361-4df9-be7d-d37aca4cbb59	uploads/feb23f1e-4f87-42b1-ae3e-c6d3eadcff58.webp	2024-11-26 17:49:03.025242+01
6e5dc6c0-c0de-490a-ba4d-3e4b035eb741	93b5307f-9c97-4fd5-a5c8-a1a10f107269	uploads/44789d56-436c-44b2-a550-5a7759b1fe74.webp	2024-11-26 18:02:20.504667+01
ed8170b3-1781-4080-9e28-b6a11f339d1a	af727efb-b517-45fc-80bb-1b8ed6c9f20a	uploads/de0cc683-9530-45f2-bafe-ce3fdff9efd4.jpg	2024-11-27 12:08:11.531187+01
f3a735c1-d434-4d7e-9373-6c788db0981c	2ab78fc3-29c7-4f07-9b0d-14678397b4a6	uploads/9f2aa896-6a28-440b-9ac4-21d440764706.webp	2024-11-27 12:12:20.181535+01
\.


--
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.items (id, title, description, price, size, category, status, quantity, seller_id, created_at) FROM stdin;
55020891-ded6-4262-bebc-36f5c7f7bbd6	K-way	Keep it yellow and mellow.	187.00	M	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:20:24.222401+01
251a1dd9-3506-4fe6-9416-cd281c89de77	K-way	Blue is true.	165.00	L	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:27:30.152+01
177afae1-a840-4f11-996e-e9f3c8378694	K-way	Green looks clean.	132.00	L	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:44:04.063173+01
7e52385e-4361-4df9-be7d-d37aca4cbb59	K-way	Purple rhymes with turtle.	400.00	S	tops	available	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:49:03.025242+01
93b5307f-9c97-4fd5-a5c8-a1a10f107269	K-way	Back in black.	128.00	S	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 18:02:20.504667+01
af727efb-b517-45fc-80bb-1b8ed6c9f20a	K-way	Pink doesn't stink.	99.00	L	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-27 12:08:11.531187+01
2ab78fc3-29c7-4f07-9b0d-14678397b4a6	K-way	Blue for you.	224.00	S	tops	available	1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-27 12:12:20.181535+01
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.order_items (id, order_id, item_id, price_at_time, created_at) FROM stdin;
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.orders (id, user_id, total, status, address, created_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.users (id, name, email, password_hash, created_at) FROM stdin;
00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Killian	killian.ledoucen@yahoo.fr	$2a$14$/E9ImhDMFxV0fVYZ.egy2evCe9e.UvH8oBAlIYlMbAUC3BZCzdjkC	2024-11-26 17:19:48.935842+01
7ad46201-3c51-4178-9dc7-29390c5044f6	Coralie	coralie_jacquier@hotmail.fr	$2a$14$EsZNa1q29LksvkEu5h3Spui9e/E9Ehu4FCIxE/3arfZqdGfpt283y	2024-11-26 22:13:15.299577+01
\.


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: cart_items cart_items_user_id_item_id_key; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_user_id_item_id_key UNIQUE (user_id, item_id);


--
-- Name: item_images item_images_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.item_images
    ADD CONSTRAINT item_images_pkey PRIMARY KEY (id);


--
-- Name: items items_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.items
    ADD CONSTRAINT items_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_cart_user; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_cart_user ON public.cart_items USING btree (user_id);


--
-- Name: idx_item_images_item; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_item_images_item ON public.item_images USING btree (item_id);


--
-- Name: idx_items_category; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_items_category ON public.items USING btree (category);


--
-- Name: idx_items_seller; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_items_seller ON public.items USING btree (seller_id);


--
-- Name: idx_orders_user; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_orders_user ON public.orders USING btree (user_id);


--
-- Name: cart_items cart_items_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: item_images item_images_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.item_images
    ADD CONSTRAINT item_images_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id) ON DELETE CASCADE;


--
-- Name: items items_seller_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.items
    ADD CONSTRAINT items_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES public.users(id);


--
-- Name: order_items order_items_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id);


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: orders orders_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

