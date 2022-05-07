-- public.users definition

CREATE TABLE public.users (
	id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
	email varchar NOT NULL,
	"password" varchar NOT NULL,
	"name" varchar NOT NULL,
	created timestamp without time zone NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT users_un UNIQUE (email)
);


