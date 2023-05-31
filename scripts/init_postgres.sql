DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  password VARCHAR NOT NULL,
  is_admin BOOLEAN NOT NULL DEFAULT false,
  provider SMALLINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT users_username_unique UNIQUE (username)
);

DROP TABLE IF EXISTS sessions;
CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT sessions_token_unique UNIQUE (token)
);

DROP TABLE IF EXISTS assignments;
CREATE TABLE assignments (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL UNIQUE,
  description VARCHAR NOT NULL,
  grader_url VARCHAR(255) NOT NULL,
  container VARCHAR(255) NOT NULL,
  part_id VARCHAR(255) NOT NULL,
  files TEXT[] NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT assignments_title_unique UNIQUE (title)
);

INSERT INTO assignments (title, description, grader_url, container, part_id, files)
  VALUES (
    'Grader #1',
    'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
    'http://runner:8021/api/v1/grader',
    'golangcourse_final',
    'HW1_game',
    '{"main.go", "lib.go"}'
  );

DROP TABLE IF EXISTS submissions;
CREATE TABLE submissions (
  id SERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  assignment_id BIGINT REFERENCES assignments(id) ON DELETE SET NULL,
  status SMALLINT NOT NULL DEFAULT 0,
  details VARCHAR,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS submission_attachments;
CREATE TABLE submission_attachments (
  id SERIAL PRIMARY KEY,
  url VARCHAR NOT NULL,
  name VARCHAR(255) NOT NULL,
  submission_id BIGINT REFERENCES submissions(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
