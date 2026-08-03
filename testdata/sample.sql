-- A line comment.
/* And a block comment
   that spans lines. */
create table notes (
	id bigint generated always as identity primary key,
	"order" int not null default 0,
	body text not null,
	created_at timestamptz not null default now()
);

create index notes_created_at on notes (created_at desc);

insert into notes (body) values ('a string
that spans lines');

select id, body
from notes
where created_at > now() - interval '7 days'
order by created_at desc
limit 10;
