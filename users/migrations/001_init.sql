-- users
create table if not exists users (
  id            bigserial primary key,
  email         citext unique not null,
  password_hash text not null,
  role          text not null default 'participant', -- participant|jury|moderator|admin|organizer
  nick          text not null,
  avatar_url    text,
  bio           text check (char_length(bio) <= 200),
  links         jsonb not null default '[]'::jsonb, -- array of URLs, <=5 (валидируем в коде)
  followers_count integer not null default 0,
  following_count integer not null default 0,
  is_banned     boolean not null default false,
  email_verified boolean not null default false,
  created_at    timestamptz not null default now()
);

-- index for login
create index if not exists idx_users_email on users (email);

-- email verify tokens
create table if not exists email_verify_tokens (
  token        text primary key,
  user_id      bigint not null references users(id) on delete cascade,
  expires_at   timestamptz not null,
  created_at   timestamptz not null default now()
);
create index if not exists idx_email_verify_expires on email_verify_tokens (expires_at);

-- for /profile/submissions (пока заглушка: только схема для пагинации)
create table if not exists jam_submissions (
  id           bigserial primary key,
  user_id      bigint not null references users(id) on delete cascade,
  jam          text not null,
  game         text not null,
  status       text not null,
  submitted_at timestamptz not null default now()
);
create index if not exists idx_submissions_user on jam_submissions (user_id, submitted_at desc);
