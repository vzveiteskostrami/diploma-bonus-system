CREATE TABLE IF NOT EXISTS 
      UDATA(
            USERID bigint not null,
            USER_NAME character varying(64) NOT NULL,
            USER_PWD character varying(64) NOT NULL,
            DELETE_FLAG boolean DEFAULT false
           );
CREATE UNIQUE INDEX IF NOT EXISTS udata_unique_on_userid ON udata (USERID);
CREATE INDEX IF NOT EXISTS udata_on_user_name ON udata (USER_NAME);

CREATE TABLE IF NOT EXISTS 
      ORDERS(
             OID bigint not null,
             USERID bigint not null,
             NUMBER character varying(64) NOT NULL,
             STATUS smallint NOT NULL,
             ACCRUAL double precision NOT NULL,
             ACCRUAL_DATE timestamp with time zone NOT NULL,
             DELETE_FLAG boolean DEFAULT false
            );
CREATE UNIQUE INDEX IF NOT EXISTS orders_unique_on_oid ON orders (OID);
CREATE INDEX IF NOT EXISTS orders_on_userid ON orders (USERID);
CREATE UNIQUE INDEX IF NOT EXISTS orders_unique_on_number ON orders (NUMBER);

create sequence if not exists gen_oid as bigint minvalue 1 no maxvalue start 1 no cycle;
