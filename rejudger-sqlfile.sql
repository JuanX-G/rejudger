--
-- PostgreSQL database dump
--

\restrict dD0tBu0mAz6dU2xL77zjgnCcEafoveaeHMZTdOynTcD18QtBQAkfjaj7Fr9r7r9

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.6

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
-- Data for Name: contexts; Type: TABLE DATA; Schema: public; Owner: macia
--

COPY public.contexts (id, name) FROM stdin;
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.permissions (id, action, context) FROM stdin;
4	View	peak
5	Submit	peak
6	Write	peak
7	PlusOne	peak
\.


--
-- Data for Name: pipelines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipelines (id, name, version, description) FROM stdin;
19	pipeline1	{"name": "pipeline1", "stages": [{"name": "review", "blind": false, "deadline": "", "soft_veto": false, "has_deadline": false, "plus_ones_required": 1}, {"name": "re review", "blind": false, "deadline": "", "soft_veto": true, "has_deadline": false, "plus_ones_required": 2}], "description": "the first pipeline"}	the first pipeline
\.


--
-- Data for Name: stages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stages (id, name, version, plus_one_required, blind, soft_veto, has_deadline, deadline_str, final) FROM stdin;
7	review	{"name": "review", "blind": false, "deadline": "", "soft_veto": false, "has_deadline": false, "plus_ones_required": 1}	1	f	f	f	\N	f
8	re review	{"name": "re review", "blind": false, "deadline": "", "soft_veto": true, "has_deadline": false, "plus_ones_required": 2}	2	f	t	f	\N	t
\.


--
-- Data for Name: pipeline_stages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipeline_stages (id, pipeline_id, stage_id, "position") FROM stdin;
37	19	7	0
38	19	8	1
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, name, internal_id, password, email) FROM stdin;
1	jm	\N	$argon2id$v=19$m=131072,t=1,p=2$i88Y/b/PKUA9iYEsr1yNIZGU2ZyU4fTVUxFyI7RXa1c$eFlm4CCFiVfxqy6TJI7OoZL26F+tVlgwm+oGRaofslI	m@m.m
\.


--
-- Data for Name: submissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.submissions (id, created_at, content, author, hash, pipeline, pipeline_stage) FROM stdin;
7	2026-11-05 20:24:07.422545	this is my tech article!	1	H5yVNbK/8kmEXuFiyOvJH6N55aBqVYVjnHI2BmqvXvEHyUGpZUFFL3CLt3frhsUBOEc/i/Sj4TOL1VAGPxf+uA==	19	38
\.


--
-- Data for Name: plus_ones; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.plus_ones (user_id, submission_id, pipeline_stage, made_at) FROM stdin;
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.roles (id, name) FROM stdin;
70	role rolity
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role_permissions (role_id, permission_id) FROM stdin;
70	5
70	6
70	7
70	4
\.


--
-- Data for Name: stage_deadlines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stage_deadlines (id, stage_id, last_ran, duration) FROM stdin;
\.


--
-- Data for Name: submission_artifacts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.submission_artifacts (id, hash, name, generation, submission_id, file_type, uploaded_at) FROM stdin;
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_roles (user_id, role_id) FROM stdin;
1	70
\.


--
-- Data for Name: vetoes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.vetoes (id, vetoer, submission_id, hard) FROM stdin;
\.


--
-- Name: contexts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: macia
--

SELECT pg_catalog.setval('public.contexts_id_seq', 1, false);


--
-- Name: permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.permissions_id_seq', 216, true);


--
-- Name: pipeline_stages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pipeline_stages_id_seq', 38, true);


--
-- Name: pipelines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pipelines_id_seq', 19, true);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.roles_id_seq', 95, true);


--
-- Name: stage_deadlines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stage_deadlines_id_seq', 1, false);


--
-- Name: stages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stages_id_seq', 8, true);


--
-- Name: submission_artifacts_generation_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.submission_artifacts_generation_seq', 1, false);


--
-- Name: submission_artifacts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.submission_artifacts_id_seq', 1, false);


--
-- Name: submissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.submissions_id_seq', 7, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 1, true);


--
-- Name: vetoes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.vetoes_id_seq', 1, false);


--
-- PostgreSQL database dump complete
--

\unrestrict dD0tBu0mAz6dU2xL77zjgnCcEafoveaeHMZTdOynTcD18QtBQAkfjaj7Fr9r7r9

