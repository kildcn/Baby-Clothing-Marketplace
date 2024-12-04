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

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: killian
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO killian;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: addresses; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.addresses (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    street text NOT NULL,
    city text NOT NULL,
    state text NOT NULL,
    zip_code text NOT NULL,
    country text NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL
);


ALTER TABLE public.addresses OWNER TO killian;

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
-- Name: message_seen; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.message_seen (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    message_id uuid,
    user_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.message_seen OWNER TO killian;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    order_id uuid NOT NULL,
    sender_id uuid NOT NULL,
    message text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_message_not_empty CHECK ((message <> ''::text))
);


ALTER TABLE public.messages OWNER TO killian;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    reference_id uuid NOT NULL,
    message text NOT NULL,
    read boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.notifications OWNER TO killian;

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
-- Name: order_status_history; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.order_status_history (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    order_id uuid,
    status public.order_status_enum NOT NULL,
    message text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.order_status_history OWNER TO killian;

--
-- Name: orders; Type: TABLE; Schema: public; Owner: killian
--

CREATE TABLE public.orders (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    address_id uuid,
    total numeric(10,2) NOT NULL,
    status public.order_status_enum DEFAULT 'pending'::public.order_status_enum,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
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
-- Data for Name: addresses; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.addresses (id, user_id, street, city, state, zip_code, country, is_default, created_at, first_name, last_name) FROM stdin;
8038a3d2-2818-49b4-9fb0-761d0f447387	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	f	2024-11-27 16:19:00.882883+01	Unknown	Unknown
a52cec1d-26b8-4579-a5b1-765e3eb9ba30	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin - Neukölln	Berlin	12055	Germany	f	2024-11-28 21:33:23.546799+01	Unknown	Unknown
4a2aee4b-0046-4c6f-9af6-6fa69ac4ba97	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	f	2024-12-02 11:10:29.660555+01		
6e12b5ed-856e-48b0-b380-9be49422e3ac	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	f	2024-12-02 15:16:55.457271+01		
eff729a8-7852-4814-a0c5-f771daf23dab	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	f	2024-12-02 15:26:32.935745+01		
b3a953e5-a7c5-497c-b16b-cf662e6698c7	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	t	2024-12-02 15:44:17.206784+01		
6967102a-2c2b-408f-9039-429a99907765	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	RichardStrasse 63	Berlin	Berlin		Germany	t	2024-12-02 15:53:16.258563+01		
8c3f9e44-cdf5-4e95-afed-11a66acb9f8c	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin		Germany	t	2024-12-03 22:00:23.502116+01		
342461b0-76c7-4ee2-b4ac-c90ccac8a89b	7ad46201-3c51-4178-9dc7-29390c5044f6	RichardStrasse 63	Berlin	Berlin	12055	Germany	t	2024-12-03 22:09:21.272255+01	Eno	Jacquier
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.cart_items (id, user_id, item_id, added_at) FROM stdin;
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
ea103f4f-0313-4804-8d9a-663a7514ba6c	ff8442c2-0e62-481a-b816-f99254480e9f	uploads/8b702165-07d5-405b-80cc-ab2fb8d438bc.jpg	2024-12-02 15:52:51.844276+01
f52b0176-f657-4963-a55c-adca2ddc50ec	6018dac7-5d87-4374-a057-e8b56bdae1a8	uploads/134ec90d-6f64-4ac8-ad43-60f25624c8e5.jpg	2024-12-03 21:59:24.600164+01
ff14ff69-a6af-442a-904d-d6a39da22fff	dd232ffb-2aab-454b-b957-cd0e390ea210	uploads/1f18b297-c00e-48cb-9fc2-9085ca968937.jpg	2024-12-03 22:08:54.803924+01
\.


--
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.items (id, title, description, price, size, category, status, quantity, seller_id, created_at) FROM stdin;
7e52385e-4361-4df9-be7d-d37aca4cbb59	K-way	Purple rhymes with turtle.	400.00	S	tops	available	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:49:03.025242+01
2ab78fc3-29c7-4f07-9b0d-14678397b4a6	K-way	Blue for you.	224.00	S	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-27 12:12:20.181535+01
177afae1-a840-4f11-996e-e9f3c8378694	K-way	Green looks clean.	132.00	L	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:44:04.063173+01
93b5307f-9c97-4fd5-a5c8-a1a10f107269	K-way	Back in black.	128.00	S	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 18:02:20.504667+01
251a1dd9-3506-4fe6-9416-cd281c89de77	K-way	Blue is true.	165.00	L	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:27:30.152+01
55020891-ded6-4262-bebc-36f5c7f7bbd6	K-way	Keep it yellow and mellow.	187.00	M	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-26 17:20:24.222401+01
af727efb-b517-45fc-80bb-1b8ed6c9f20a	K-way	Pink doesn't stink.	99.00	L	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-27 12:08:11.531187+01
ff8442c2-0e62-481a-b816-f99254480e9f	K-way	Looks good right?	444.00	M	tops	sold	0	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-12-02 15:52:51.844276+01
6018dac7-5d87-4374-a057-e8b56bdae1a8	k-way	black	555.00	XS	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-03 21:59:24.600164+01
dd232ffb-2aab-454b-b957-cd0e390ea210	k-way	black black black	999.00	XL	tops	sold	0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-03 22:08:54.803924+01
\.


--
-- Data for Name: message_seen; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.message_seen (id, message_id, user_id, created_at) FROM stdin;
c6cae366-90d4-4514-b600-07a018242681	f53e5a45-bddc-4c30-8d01-367d333eb166	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
d2579f58-5d4a-4a32-a114-562255a9d21e	897ca9bf-2b4f-4751-9ef5-37d7c5887e4b	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
b547325b-93b9-4e77-81a0-c189ee6df938	0dc4f51f-fb84-4ea5-84f3-ff6ee401eac8	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
0b88acd2-1757-4fc9-b016-0f0713e14117	cb406694-c006-43c8-ae1a-0bff5502999e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
979cbc24-6bbd-4183-ba37-03352873b592	d3b86c0c-8960-47a6-99f4-2f357a4aa106	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
547ba96f-ca60-4a91-b3d4-05eb5f78bfd2	2655838c-fecd-47c6-8e34-5c18b3ec1098	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
81969b41-f62b-4ff7-91b2-66adaca6a964	312a523c-f591-4227-98fd-5ada6147a2f4	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
f64b1a25-4742-405f-9394-1daf847aa280	5568c0c8-26b5-4177-bf27-0df3bee40848	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
15420ca2-1601-4300-84a6-374083a0256c	474aaec0-a3fa-42be-8f6a-4542ac0fe119	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
dbc7febc-c9e4-4960-8a18-c994de2a6bc1	4ea16e67-b622-4a63-9afc-54a2206ad3fb	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
8a148efc-88cd-4dd3-8658-cc51b022797d	cc09ad2e-2e23-47b1-9517-a641e1a26c30	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
8b182b65-3f5c-4652-934d-c5e667edcbbc	8a608c48-52b7-4307-92b8-d0340d7a5533	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
78d6ce19-b54c-472c-b8ee-659c175209f6	50667af7-7887-4df5-b6e1-5c320ed67848	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
2bb686e3-ec4b-4a50-93eb-b483812c47b0	88d0d6d8-2d02-40a3-9873-4a0146468634	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
11838438-2c9a-4123-9cc6-cff79ea642ba	c17531e5-d431-4de3-93a5-01daed0d60d5	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
a0fb2ae5-7f94-4466-bab9-c6ddb51e8720	48ac3da7-fb43-425e-80b5-824c72159037	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
46dbb550-30cd-49e0-893b-4dbd76d24fdc	6b524842-7c30-44c9-9981-2f13bcfe4a49	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
6236cf60-4472-4678-8f36-bf28f94ff0e5	1964fe19-9e33-4f57-acf9-a47fab2a8b3f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
7bbb98d2-7210-4973-84a4-6bda08ed7ac9	0325d670-8c51-40b3-831c-309b66aed868	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
6a6a5eba-b344-489e-bcd2-f72b6f853f91	a64ce444-3952-4112-a1c3-91adfd3ffc29	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
f5da9d49-052d-4f40-afa1-cb4a268d2b72	90c6c407-7edc-41af-bdb2-436b004f314b	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
18781bd5-ff55-47c4-bb2a-456012ac74fc	ea3feade-8b9e-4f53-986e-4dc4c454854e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
0f89661f-c339-4a71-8d6d-c8d39ae639ad	50d9408b-c0ee-4f62-8c56-b4625d414b2e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
113b9b85-dc8c-44e1-9766-1db8624a526b	684586f2-fe0d-4771-b3ca-1ce99f7697d5	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
f57b69be-b6f7-4a27-97b7-d91d41c2b45f	c35e8c00-11d9-4c7d-87ef-8e6331b1a527	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
1f06cc01-0b11-403c-889c-eb08d0f1b3c6	bf533a44-5c26-43a4-9334-d16f3a0b2625	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
5497a333-de94-4b9c-ac7b-8b6e308903ca	aa9e0ade-3457-48df-80bc-bf8cc99f6c25	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
aac83087-e1e8-414f-983a-5bc11b88a486	a031ca27-50da-4b68-95e4-570c98c9e453	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
cd7ec86b-d432-40e7-9b76-2383bf36df43	dd1e53d3-17e4-48aa-9b5e-368e6d31578d	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
dbd43049-b2f1-4fb3-a4f0-a29a170c776f	f3ed1ffe-46d9-43a4-bae4-ec6ee4560b7f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
cc2f87eb-3e60-4744-b069-56ab2186f838	4ba0b568-1b59-4671-a2a8-4e9e88f933b2	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
85080381-c243-4e07-a65a-e9e87047674d	76d4dfea-0423-45ab-8000-db9f2b97bf8e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
35751498-9407-4d9a-b0a3-36c870b16b3a	5f5726af-bbf7-4bfc-926f-b41c61144600	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
388529ab-3510-450b-a78c-a416756b8880	3cb552f9-0de3-4eb5-876c-77aa92861641	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
557956f6-b18f-4e4c-ab43-2722e5738e60	ad2b9987-ea44-4bf3-a49f-d5bb823d50a5	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
2511338d-4416-4ed7-8241-2343901f3ab6	a657643b-2a59-4c65-998d-8d49a48d4da5	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
915cfabf-ec90-4771-9301-214580eb42fd	920a7497-145f-4001-b809-5927c28e4791	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
fb16ddd6-2a8b-4cda-83c6-b01195586407	0ccd9e3a-1c8a-4ca7-9b98-648934f8314a	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
ac8cee88-dca0-4856-b883-3759f1539be8	0a2d0c64-23be-4bbd-9d5c-7fa9aaeb922f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
05464cd4-cda0-4e65-8c8f-b1ca83272cb3	559e5cff-1a7c-4652-9a94-e481e5e43032	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
162e65dd-2715-4f25-9734-265172b72f31	dd0ba9cd-a000-4226-ad43-7b79a2b0c58f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
7d256ecf-7f5d-49a2-a77d-d27418a2dc42	82b2f93a-843d-440e-960e-5cdc3e7f4535	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
e734f1bb-e583-4516-b02d-9e5fb9008cba	f7761d5e-5961-4f6b-bea3-ebeb68839f0f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
39fbec8a-b49e-4cce-bdeb-2857d512489b	eeaab26b-0bba-456c-a690-045136f8181a	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
8fc4a3ca-31ae-4a15-996a-f680b20a9540	bac0c107-d9ab-4082-97a4-e68284bcfd31	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
ceea34d4-8a2c-4c6d-9748-473aa7b289aa	5f7eae1f-9a87-42d8-9f0c-5ca9e1783fb9	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
e29f36b8-aa86-4e81-9303-509190fef148	2d6ba99e-9a0c-40f7-a864-e7f14daf3635	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
3df745b0-b1ca-44db-9f55-73ae6488e5eb	81b09091-188f-4fb8-bd9b-9fa74f19e5c7	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
56993776-5033-42e9-84e3-6f97f9147a8b	3e2f808b-28a3-4d0c-ad46-2e336f462a68	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
53f280ad-19df-4060-a6e2-f2137ce41236	00b9fff6-1484-4780-8f9d-5d2a81c3076f	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
63f6ca9c-bb0a-4e57-bdb9-89e4f5ba9d4a	f4a5c2e5-97bb-40ca-80d4-0438f9529602	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:02:42.702477+01
b3a26d87-4cc9-4b4a-a7c5-74626b2aa30b	f53e5a45-bddc-4c30-8d01-367d333eb166	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
aecf5a0b-e4f1-4737-9950-f03120b71212	897ca9bf-2b4f-4751-9ef5-37d7c5887e4b	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
882d6777-941d-440f-b003-229ef643eed9	0dc4f51f-fb84-4ea5-84f3-ff6ee401eac8	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
c07455b7-db13-4d99-aa0a-3d72e251e690	cb406694-c006-43c8-ae1a-0bff5502999e	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
5ae85410-6aa2-4c9c-b1ae-bf6d690f1671	d3b86c0c-8960-47a6-99f4-2f357a4aa106	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
8fd9adc4-8ffb-4f76-8a63-7b561158332a	2655838c-fecd-47c6-8e34-5c18b3ec1098	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
618d37c4-f2dd-4b51-9ed4-3813502f6b9d	312a523c-f591-4227-98fd-5ada6147a2f4	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
30579b42-6a0a-424c-8a8d-79c173ea7cf2	5568c0c8-26b5-4177-bf27-0df3bee40848	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
2f15b17d-da12-480b-a76a-297c2371bd21	474aaec0-a3fa-42be-8f6a-4542ac0fe119	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
182de3f2-02e8-457a-981c-c73ba7e476c9	4ea16e67-b622-4a63-9afc-54a2206ad3fb	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
744ae3ee-f2ca-4fb7-aedd-d008c0a3fd86	cc09ad2e-2e23-47b1-9517-a641e1a26c30	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
2796435f-ade6-40f8-b35e-284858442913	8a608c48-52b7-4307-92b8-d0340d7a5533	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
59598fb2-12d9-4c69-afc5-957c1cf52e4c	50667af7-7887-4df5-b6e1-5c320ed67848	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
13b6efaf-de8c-4dfe-9a42-163c908f2102	88d0d6d8-2d02-40a3-9873-4a0146468634	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
74909031-641f-4818-9930-deca779ae887	c17531e5-d431-4de3-93a5-01daed0d60d5	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
39e8d890-7ce5-40a9-970f-41571bb6bd8e	48ac3da7-fb43-425e-80b5-824c72159037	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
ddf35430-991b-4f34-b089-76f768fea52f	6b524842-7c30-44c9-9981-2f13bcfe4a49	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
41e27269-0f52-437b-ace7-749cd0134347	1964fe19-9e33-4f57-acf9-a47fab2a8b3f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
2a7d22a1-88ed-472c-b651-6ad066fdd83a	0325d670-8c51-40b3-831c-309b66aed868	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
ef58fa6e-f17e-474d-af8e-5921724987ee	a64ce444-3952-4112-a1c3-91adfd3ffc29	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
b927075d-ca33-4e8d-ad24-edd113a858ee	90c6c407-7edc-41af-bdb2-436b004f314b	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
8ebcc62f-0c13-470e-abb9-15db4a41ba9f	ea3feade-8b9e-4f53-986e-4dc4c454854e	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
06b945ed-5f81-476b-9d38-ff3007f0a4b4	50d9408b-c0ee-4f62-8c56-b4625d414b2e	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
bad064bb-7b85-471d-8952-b16699ad447c	684586f2-fe0d-4771-b3ca-1ce99f7697d5	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
3a3bf328-108c-44d6-b8b7-5540b11013c7	c35e8c00-11d9-4c7d-87ef-8e6331b1a527	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
fe0e9871-7f12-470c-a18a-9a12b40ceb05	bf533a44-5c26-43a4-9334-d16f3a0b2625	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
7a958525-071f-45e4-805e-5cfc0f527d49	aa9e0ade-3457-48df-80bc-bf8cc99f6c25	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
bb4be2b2-1a71-4c1c-9906-2b260ef3c2f4	a031ca27-50da-4b68-95e4-570c98c9e453	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
799b9fca-f6c9-45a9-97ec-a47206d61005	dd1e53d3-17e4-48aa-9b5e-368e6d31578d	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
17bc2633-8433-48cb-8308-270206be75c0	f3ed1ffe-46d9-43a4-bae4-ec6ee4560b7f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
d11b046a-0216-42e8-b40a-001a390bc0b5	4ba0b568-1b59-4671-a2a8-4e9e88f933b2	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
c9fb2207-2c95-495e-b0ea-3b9d79399d21	76d4dfea-0423-45ab-8000-db9f2b97bf8e	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
db1ad286-6f56-46ee-8fb9-cf121db035b7	5f5726af-bbf7-4bfc-926f-b41c61144600	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
0a0c744a-04df-4042-b8b1-02c3969a19a3	3cb552f9-0de3-4eb5-876c-77aa92861641	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
d0360f6b-c9dc-4241-97fa-16edfa1ff704	ad2b9987-ea44-4bf3-a49f-d5bb823d50a5	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
82ff815e-6143-4ddd-ae65-1b19bfde737c	a657643b-2a59-4c65-998d-8d49a48d4da5	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
66774683-802e-416d-9fb6-ec87b20fcbbe	920a7497-145f-4001-b809-5927c28e4791	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
a66e52df-eb72-4f86-bb91-674d09755275	0ccd9e3a-1c8a-4ca7-9b98-648934f8314a	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
2848e2c7-4eab-4f5e-a6bf-6e346353d9b4	0a2d0c64-23be-4bbd-9d5c-7fa9aaeb922f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
fb98c3be-917e-4e3a-a14a-88f89049d35d	559e5cff-1a7c-4652-9a94-e481e5e43032	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
4ec4201b-f91f-4998-a995-40d5822b0001	dd0ba9cd-a000-4226-ad43-7b79a2b0c58f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
12fba10a-cf52-4537-a9f0-674bc3475f58	82b2f93a-843d-440e-960e-5cdc3e7f4535	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
8c88dcba-497e-4268-98ec-b36f60b985e3	f7761d5e-5961-4f6b-bea3-ebeb68839f0f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
ddea4716-d309-486c-a79d-ed308ade4bfd	eeaab26b-0bba-456c-a690-045136f8181a	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
e5792552-cbf2-4b77-bf91-d368d201b50e	bac0c107-d9ab-4082-97a4-e68284bcfd31	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
db9a0b25-28c7-4c0e-a6d1-6d015cab3ebe	5f7eae1f-9a87-42d8-9f0c-5ca9e1783fb9	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
fc14809d-4306-43d3-a73b-f3dc86d4ac38	2d6ba99e-9a0c-40f7-a864-e7f14daf3635	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
2a64a185-63c7-46c3-af6a-af97b713f427	81b09091-188f-4fb8-bd9b-9fa74f19e5c7	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
b9f3836f-b9c3-4a0c-839c-5f33056e9bdd	3e2f808b-28a3-4d0c-ad46-2e336f462a68	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
911bdeab-a4bf-4daf-bc5b-8018b6f178ea	00b9fff6-1484-4780-8f9d-5d2a81c3076f	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
b8c12d1b-ae74-4441-94b8-915b28a5abfa	f4a5c2e5-97bb-40ca-80d4-0438f9529602	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
c919cabd-7e50-43da-99ed-78939e75ad3c	2c30a099-325f-4e0c-882b-3865680e94d0	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:03:11.917579+01
fc2d3b3c-f9ed-411b-92f1-ab9c5d3a7b30	2c30a099-325f-4e0c-882b-3865680e94d0	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:03:51.381736+01
7ca29ffd-70a8-42fc-97f7-ade3fceb0da8	58323a6f-a0c4-47d0-a59b-8c0ec4dd5da9	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:03:51.381736+01
c25ce7c6-4f7a-4149-a2b8-6aae71fe6597	58323a6f-a0c4-47d0-a59b-8c0ec4dd5da9	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:10:24.561543+01
0e2bc5ca-ffba-4bf3-807e-77185ffa8cfb	5c520595-cc9e-43fb-84a9-fa851d77799e	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-11-28 15:10:24.561543+01
928d8224-89f5-4ce1-8428-abe0fb4a0cfc	5c520595-cc9e-43fb-84a9-fa851d77799e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:48:27.156465+01
de54339d-5cf0-482f-8e33-81113d467eb9	43fc3d56-d2c2-43ce-bd8a-dc278076735c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-28 15:48:27.156465+01
6d6e5ee1-b402-4acc-938e-5f33e6000b3a	0b03fbe6-4b2c-4e99-8d7c-32ed6d5e7533	7ad46201-3c51-4178-9dc7-29390c5044f6	2024-12-02 15:18:45.226833+01
34f1e0ce-b19f-4c4b-b381-f554cd0ca79d	0b03fbe6-4b2c-4e99-8d7c-32ed6d5e7533	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-02 15:19:02.459901+01
5972425d-fc81-4cf3-8c3a-734ea82ee8f1	bdc0d0e3-9479-428c-bca4-c2e470a75d21	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-02 15:19:02.459901+01
72304dbd-42b7-4574-8dba-0c290d9005da	104fcc2d-dfa0-4066-8ac4-b105626c2c1e	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-02 16:01:02.441223+01
\.


--
-- Data for Name: messages; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.messages (id, order_id, sender_id, message, created_at) FROM stdin;
f53e5a45-bddc-4c30-8d01-367d333eb166	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Hello, thank you for your purchase!	2024-11-27 17:05:34.371929+01
897ca9bf-2b4f-4751-9ef5-37d7c5887e4b	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Hmmm.. Thank you but you are the one who purchased!	2024-11-27 17:06:23.871916+01
0dc4f51f-fb84-4ea5-84f3-ff6ee401eac8	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	LOL my bad... Thank for sending it! Can't wait to get it.	2024-11-27 17:07:54.940742+01
cb406694-c006-43c8-ae1a-0bff5502999e	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	No worries!	2024-11-27 17:09:00.01844+01
d3b86c0c-8960-47a6-99f4-2f357a4aa106	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	You'll be so stylish in the rain!	2024-11-27 21:14:37.953184+01
2655838c-fecd-47c6-8e34-5c18b3ec1098	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	I can't wait!	2024-11-27 21:16:16.441682+01
312a523c-f591-4227-98fd-5ada6147a2f4	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	How long should it take?	2024-11-27 21:35:35.773334+01
5568c0c8-26b5-4177-bf27-0df3bee40848	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	About 3-4 business days.	2024-11-27 21:36:10.838808+01
474aaec0-a3fa-42be-8f6a-4542ac0fe119	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Great!	2024-11-27 21:36:23.705491+01
4ea16e67-b622-4a63-9afc-54a2206ad3fb	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Also if you need a different size feel free to return it!	2024-11-27 21:37:57.805819+01
cc09ad2e-2e23-47b1-9517-a641e1a26c30	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Ok?	2024-11-27 21:38:19.525012+01
8a608c48-52b7-4307-92b8-d0340d7a5533	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Hellooooo?	2024-11-27 21:38:42.703178+01
50667af7-7887-4df5-b6e1-5c320ed67848	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Woops hadn't seen, of course thanks!	2024-11-27 21:39:08.074664+01
88d0d6d8-2d02-40a3-9873-4a0146468634	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Testing?	2024-11-27 21:53:41.542541+01
c17531e5-d431-4de3-93a5-01daed0d60d5	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Did you receive the item?	2024-11-28 09:44:18.920757+01
48ac3da7-fb43-425e-80b5-824c72159037	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	It should arrive today or tomorrow	2024-11-28 09:48:00.383265+01
6b524842-7c30-44c9-9981-2f13bcfe4a49	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Unless it's a bank holiday or something in Germany	2024-11-28 09:48:54.674449+01
1964fe19-9e33-4f57-acf9-a47fab2a8b3f	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Have a great day!	2024-11-28 09:53:10.98595+01
0325d670-8c51-40b3-831c-309b66aed868	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Not yet, I'll keep you posted, thank you!	2024-11-28 09:53:37.22106+01
a64ce444-3952-4112-a1c3-91adfd3ffc29	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	It's a rare find, can't wait to wear it!	2024-11-28 09:54:23.604571+01
90c6c407-7edc-41af-bdb2-436b004f314b	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	It sure is, glad you found it.	2024-11-28 09:59:35.250546+01
ea3feade-8b9e-4f53-986e-4dc4c454854e	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	:D	2024-11-28 10:00:28.188673+01
50d9408b-c0ee-4f62-8c56-b4625d414b2e	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	:D	2024-11-28 10:02:19.989231+01
684586f2-fe0d-4771-b3ca-1ce99f7697d5	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	:D	2024-11-28 10:05:04.026907+01
c35e8c00-11d9-4c7d-87ef-8e6331b1a527	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	:D	2024-11-28 10:05:25.333729+01
bf533a44-5c26-43a4-9334-d16f3a0b2625	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	:D :D	2024-11-28 10:06:45.333832+01
aa9e0ade-3457-48df-80bc-bf8cc99f6c25	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	YOYOYO	2024-11-28 10:10:05.977451+01
a031ca27-50da-4b68-95e4-570c98c9e453	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	LOOOOL	2024-11-28 10:11:40.558354+01
dd1e53d3-17e4-48aa-9b5e-368e6d31578d	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	LMAO	2024-11-28 10:11:51.292556+01
f3ed1ffe-46d9-43a4-bae4-ec6ee4560b7f	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	XD	2024-11-28 10:13:37.68934+01
4ba0b568-1b59-4671-a2a8-4e9e88f933b2	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Hilarious	2024-11-28 10:14:59.145644+01
76d4dfea-0423-45ab-8000-db9f2b97bf8e	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Fun	2024-11-28 10:16:42.486502+01
5f5726af-bbf7-4bfc-926f-b41c61144600	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Yup	2024-11-28 10:17:00.154869+01
3cb552f9-0de3-4eb5-876c-77aa92861641	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Can you hear me?	2024-11-28 11:15:49.517367+01
ad2b9987-ea44-4bf3-a49f-d5bb823d50a5	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	No	2024-11-28 11:17:16.183406+01
a657643b-2a59-4c65-998d-8d49a48d4da5	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	What's up?	2024-11-28 11:19:48.840468+01
920a7497-145f-4001-b809-5927c28e4791	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Not much, you?	2024-11-28 11:20:12.803852+01
0ccd9e3a-1c8a-4ca7-9b98-648934f8314a	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Nothing	2024-11-28 11:20:42.604923+01
0a2d0c64-23be-4bbd-9d5c-7fa9aaeb922f	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	kk	2024-11-28 11:20:55.926859+01
559e5cff-1a7c-4652-9a94-e481e5e43032	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	What you up to?	2024-11-28 11:22:00.434383+01
dd0ba9cd-a000-4226-ad43-7b79a2b0c58f	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Huh?	2024-11-28 11:22:39.009425+01
82b2f93a-843d-440e-960e-5cdc3e7f4535	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	What	2024-11-28 11:24:26.903876+01
f7761d5e-5961-4f6b-bea3-ebeb68839f0f	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	What what	2024-11-28 11:26:37.854759+01
eeaab26b-0bba-456c-a690-045136f8181a	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	What what what	2024-11-28 11:27:04.440145+01
bac0c107-d9ab-4082-97a4-e68284bcfd31	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	What you up to?	2024-11-28 11:41:15.743549+01
5f7eae1f-9a87-42d8-9f0c-5ca9e1783fb9	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Not much you?	2024-11-28 11:42:02.036723+01
2d6ba99e-9a0c-40f7-a864-e7f14daf3635	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	?	2024-11-28 11:42:12.616642+01
81b09091-188f-4fb8-bd9b-9fa74f19e5c7	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Any news?	2024-11-28 14:28:16.560143+01
3e2f808b-28a3-4d0c-ad46-2e336f462a68	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Hey	2024-11-28 14:41:56.674223+01
00b9fff6-1484-4780-8f9d-5d2a81c3076f	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	ho	2024-11-28 14:48:19.254628+01
f4a5c2e5-97bb-40ca-80d4-0438f9529602	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	let's go!	2024-11-28 14:49:29.76417+01
2c30a099-325f-4e0c-882b-3865680e94d0	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	LOLLLL	2024-11-28 15:02:50.616645+01
58323a6f-a0c4-47d0-a59b-8c0ec4dd5da9	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	LOLOLOL	2024-11-28 15:03:28.623514+01
5c520595-cc9e-43fb-84a9-fa851d77799e	7df53cd6-3f58-4c9d-a080-be223412582c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Awesome	2024-11-28 15:10:10.225673+01
43fc3d56-d2c2-43ce-bd8a-dc278076735c	7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	Ground control?	2024-11-28 15:47:46.327891+01
0b03fbe6-4b2c-4e99-8d7c-32ed6d5e7533	ce8dda13-bfab-4d19-8a86-02458be6ffe4	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Just making sure this wasn't an accident?	2024-12-02 15:18:32.544295+01
bdc0d0e3-9479-428c-bca4-c2e470a75d21	ce8dda13-bfab-4d19-8a86-02458be6ffe4	7ad46201-3c51-4178-9dc7-29390c5044f6	No all good!	2024-12-02 15:18:52.127726+01
104fcc2d-dfa0-4066-8ac4-b105626c2c1e	167a297a-5853-41a6-8c21-2fc7fc4a2ec4	7ad46201-3c51-4178-9dc7-29390c5044f6	Hey	2024-12-02 16:00:50.70785+01
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.notifications (id, user_id, type, reference_id, message, read, created_at) FROM stdin;
a54b8f57-c588-4e42-825b-df15cbedac5b	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	ea634696-6f67-4a5e-bb29-e3c586a86aff	New order #ea634696-6f67-4a5e-bb29-e3c586a86aff received	t	2024-12-02 15:26:32.941754+01
2a416f82-dd5d-4e64-948f-bbee7316754c	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	ce8dda13-bfab-4d19-8a86-02458be6ffe4	New order #ce8dda13-bfab-4d19-8a86-02458be6ffe4 received	t	2024-12-02 15:16:55.460681+01
75a0bd81-16a4-4ad2-891b-0447fc8a5fe6	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	3c9985a2-ba91-4950-a839-935ae18432bc	New order #3c9985a2-ba91-4950-a839-935ae18432bc received	t	2024-12-02 11:10:29.666961+01
7a2638b8-f771-48f2-bb45-5c884e8fd84a	7ad46201-3c51-4178-9dc7-29390c5044f6	order_status	3c9985a2-ba91-4950-a839-935ae18432bc	Your order #3c9985a2-ba91-4950-a839-935ae18432bc has been cancelled. Reason: Order cancelled. Reason: Wrong size	t	2024-12-02 11:13:30.160944+01
373933da-5a69-4015-8df6-55bfb2e81c8b	7ad46201-3c51-4178-9dc7-29390c5044f6	order_status	ce8dda13-bfab-4d19-8a86-02458be6ffe4	Your order #ce8dda13-bfab-4d19-8a86-02458be6ffe4 has been shipped	t	2024-12-02 15:19:17.078456+01
30ea71b4-591e-4a12-9746-eb29a10cc4e0	7ad46201-3c51-4178-9dc7-29390c5044f6	order_status	ea634696-6f67-4a5e-bb29-e3c586a86aff	Your order #ea634696-6f67-4a5e-bb29-e3c586a86aff has been shipped	t	2024-12-02 15:30:48.271306+01
000feaf6-d993-43c1-bfca-0ee7b24ac47d	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	167a297a-5853-41a6-8c21-2fc7fc4a2ec4	New order #167a297a-5853-41a6-8c21-2fc7fc4a2ec4 received	t	2024-12-02 15:44:17.211173+01
02db023c-991a-40da-9263-21ee09684b75	7ad46201-3c51-4178-9dc7-29390c5044f6	order_status	167a297a-5853-41a6-8c21-2fc7fc4a2ec4	Your order #167a297a-5853-41a6-8c21-2fc7fc4a2ec4 has been cancelled. Reason: Order cancelled. Reason: IDK	t	2024-12-02 15:51:28.143269+01
eb5549fa-fb9e-46ec-8761-a370b819f8ac	7ad46201-3c51-4178-9dc7-29390c5044f6	order_status	77bec4e1-7835-4426-9014-cd6fede66aa1	New order #77bec4e1-7835-4426-9014-cd6fede66aa1 received	t	2024-12-02 15:53:16.261773+01
f1820243-4bcd-43c6-9361-41d49df980b5	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	48b8f114-7de6-40ce-86c6-65d83bae2201	New order #48b8f114-7de6-40ce-86c6-65d83bae2201 received	t	2024-12-03 22:00:23.505938+01
a8ad0d81-78be-4f8d-97ce-f589fedc29dd	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	f9a11f20-107f-45a0-b328-1870e2553b1d	New order #f9a11f20-107f-45a0-b328-1870e2553b1d received	t	2024-12-03 22:09:21.277362+01
35fb653c-f878-4ec8-8add-8d220dd5e55b	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	order_status	7df53cd6-3f58-4c9d-a080-be223412582c	Order #7df53cd6-3f58-4c9d-a080-be223412582c has been confirmed as delivered	f	2024-12-04 09:59:12.082283+01
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.order_items (id, order_id, item_id, price_at_time, created_at) FROM stdin;
79edfdd5-888e-4efb-9b6b-4e53e5ca8898	7df53cd6-3f58-4c9d-a080-be223412582c	2ab78fc3-29c7-4f07-9b0d-14678397b4a6	224.00	2024-11-27 16:19:00.882883+01
de347942-63fe-464a-8ff9-a9dfcf8b81a5	5856ffa9-7f8f-47b4-bd80-cde6c5145766	177afae1-a840-4f11-996e-e9f3c8378694	132.00	2024-11-28 21:33:23.546799+01
11924317-4497-42c3-be75-965aa358b011	3c9985a2-ba91-4950-a839-935ae18432bc	93b5307f-9c97-4fd5-a5c8-a1a10f107269	128.00	2024-12-02 11:10:29.660555+01
af5098c4-5255-4ebb-99d5-e69c568cc3ef	ce8dda13-bfab-4d19-8a86-02458be6ffe4	251a1dd9-3506-4fe6-9416-cd281c89de77	165.00	2024-12-02 15:16:55.457271+01
002c5fd5-211c-4571-a2b3-a0a2f67d32f4	ea634696-6f67-4a5e-bb29-e3c586a86aff	55020891-ded6-4262-bebc-36f5c7f7bbd6	187.00	2024-12-02 15:26:32.935745+01
86d2eb9a-3152-45aa-bbcd-66ab92811be7	167a297a-5853-41a6-8c21-2fc7fc4a2ec4	af727efb-b517-45fc-80bb-1b8ed6c9f20a	99.00	2024-12-02 15:44:17.206784+01
155b28a9-9abd-487e-8ef4-7a9bf39bcd4a	77bec4e1-7835-4426-9014-cd6fede66aa1	ff8442c2-0e62-481a-b816-f99254480e9f	444.00	2024-12-02 15:53:16.258563+01
ef136495-529d-46f9-a1a1-3174cb876266	48b8f114-7de6-40ce-86c6-65d83bae2201	6018dac7-5d87-4374-a057-e8b56bdae1a8	555.00	2024-12-03 22:00:23.502116+01
c00a61bc-06c6-49e1-9a55-67bc167a12b0	f9a11f20-107f-45a0-b328-1870e2553b1d	dd232ffb-2aab-454b-b957-cd0e390ea210	999.00	2024-12-03 22:09:21.272255+01
\.


--
-- Data for Name: order_status_history; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.order_status_history (id, order_id, status, message, created_by, created_at) FROM stdin;
b5c3377d-93da-45c0-92b0-5fba607804fa	7df53cd6-3f58-4c9d-a080-be223412582c	shipped	Tracking number: 2DFFFLLLLSSER331	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-11-27 17:07:05.772078+01
2e0f96c9-5e63-49ed-90ae-4f3eca0ca2df	5856ffa9-7f8f-47b4-bd80-cde6c5145766	shipped	Tracking number: 4LFAA555TL	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	2024-12-02 10:40:47.418518+01
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.orders (id, user_id, address_id, total, status, created_at, updated_at) FROM stdin;
3c9985a2-ba91-4950-a839-935ae18432bc	7ad46201-3c51-4178-9dc7-29390c5044f6	4a2aee4b-0046-4c6f-9af6-6fa69ac4ba97	128.00	cancelled	2024-12-02 11:10:29.660555+01	2024-12-02 11:13:30.157677+01
5856ffa9-7f8f-47b4-bd80-cde6c5145766	7ad46201-3c51-4178-9dc7-29390c5044f6	a52cec1d-26b8-4579-a5b1-765e3eb9ba30	132.00	delivered	2024-11-28 21:33:23.546799+01	2024-12-02 15:16:33.159411+01
ce8dda13-bfab-4d19-8a86-02458be6ffe4	7ad46201-3c51-4178-9dc7-29390c5044f6	6e12b5ed-856e-48b0-b380-9be49422e3ac	165.00	delivered	2024-12-02 15:16:55.457271+01	2024-12-02 15:24:17.444152+01
ea634696-6f67-4a5e-bb29-e3c586a86aff	7ad46201-3c51-4178-9dc7-29390c5044f6	eff729a8-7852-4814-a0c5-f771daf23dab	187.00	delivered	2024-12-02 15:26:32.935745+01	2024-12-02 15:31:40.097909+01
167a297a-5853-41a6-8c21-2fc7fc4a2ec4	7ad46201-3c51-4178-9dc7-29390c5044f6	b3a953e5-a7c5-497c-b16b-cf662e6698c7	99.00	cancelled	2024-12-02 15:44:17.206784+01	2024-12-02 15:51:28.137022+01
77bec4e1-7835-4426-9014-cd6fede66aa1	00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	6967102a-2c2b-408f-9039-429a99907765	444.00	pending	2024-12-02 15:53:16.258563+01	2024-12-02 15:53:16.258563+01
48b8f114-7de6-40ce-86c6-65d83bae2201	7ad46201-3c51-4178-9dc7-29390c5044f6	8c3f9e44-cdf5-4e95-afed-11a66acb9f8c	555.00	pending	2024-12-03 22:00:23.502116+01	2024-12-03 22:00:23.502116+01
f9a11f20-107f-45a0-b328-1870e2553b1d	7ad46201-3c51-4178-9dc7-29390c5044f6	342461b0-76c7-4ee2-b4ac-c90ccac8a89b	999.00	pending	2024-12-03 22:09:21.272255+01	2024-12-03 22:09:21.272255+01
7df53cd6-3f58-4c9d-a080-be223412582c	7ad46201-3c51-4178-9dc7-29390c5044f6	8038a3d2-2818-49b4-9fb0-761d0f447387	224.00	delivered	2024-11-27 16:19:00.882883+01	2024-12-04 09:59:12.075021+01
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: killian
--

COPY public.users (id, name, email, password_hash, created_at) FROM stdin;
00ad4f5e-bff0-4ec3-ad7b-5a04ed797f07	Killian	killian.ledoucen@yahoo.fr	$2a$14$/E9ImhDMFxV0fVYZ.egy2evCe9e.UvH8oBAlIYlMbAUC3BZCzdjkC	2024-11-26 17:19:48.935842+01
7ad46201-3c51-4178-9dc7-29390c5044f6	Coralie	coralie_jacquier@hotmail.fr	$2a$14$EsZNa1q29LksvkEu5h3Spui9e/E9Ehu4FCIxE/3arfZqdGfpt283y	2024-11-26 22:13:15.299577+01
\.


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


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
-- Name: message_seen message_seen_message_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.message_seen
    ADD CONSTRAINT message_seen_message_id_user_id_key UNIQUE (message_id, user_id);


--
-- Name: message_seen message_seen_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.message_seen
    ADD CONSTRAINT message_seen_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_status_history order_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_pkey PRIMARY KEY (id);


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
-- Name: idx_addresses_user; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_addresses_user ON public.addresses USING btree (user_id);


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
-- Name: idx_message_seen_message; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_message_seen_message ON public.message_seen USING btree (message_id);


--
-- Name: idx_message_seen_user; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_message_seen_user ON public.message_seen USING btree (user_id);


--
-- Name: idx_messages_order; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_messages_order ON public.messages USING btree (order_id);


--
-- Name: idx_messages_sender; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_messages_sender ON public.messages USING btree (sender_id);


--
-- Name: idx_notifications_created_at; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_notifications_created_at ON public.notifications USING btree (created_at);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_notifications_user_id ON public.notifications USING btree (user_id);


--
-- Name: idx_order_items_order; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_order_items_order ON public.order_items USING btree (order_id);


--
-- Name: idx_order_status_history_order; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_order_status_history_order ON public.order_status_history USING btree (order_id);


--
-- Name: idx_orders_user; Type: INDEX; Schema: public; Owner: killian
--

CREATE INDEX idx_orders_user ON public.orders USING btree (user_id);


--
-- Name: orders update_orders_updated_at; Type: TRIGGER; Schema: public; Owner: killian
--

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON public.orders FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: addresses addresses_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


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
-- Name: message_seen message_seen_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.message_seen
    ADD CONSTRAINT message_seen_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(id) ON DELETE CASCADE;


--
-- Name: message_seen message_seen_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.message_seen
    ADD CONSTRAINT message_seen_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: messages messages_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: messages messages_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.users(id);


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: order_items order_items_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id);


--
-- Name: order_status_history order_status_history_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: order_status_history order_status_history_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: orders orders_address_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_address_id_fkey FOREIGN KEY (address_id) REFERENCES public.addresses(id);


--
-- Name: orders orders_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: killian
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

