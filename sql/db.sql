CREATE TABLE public.users (
	id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
	email varchar NOT NULL,
	"password" varchar NOT NULL,
	"name" varchar NOT NULL,
	created timestamp without time zone NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT users_un UNIQUE (email)
);


CREATE TABLE public.user_sessions (
	id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
	user_id integer NOT NULL,
	"session" varchar NOT NULL,
	created time without time zone NOT NULL,
	CONSTRAINT user_sessions_pk PRIMARY KEY (id),
	CONSTRAINT user_sessions_un UNIQUE ("session"),
	CONSTRAINT user_sessions_fk FOREIGN KEY (user_id) REFERENCES public.users(id)
);

