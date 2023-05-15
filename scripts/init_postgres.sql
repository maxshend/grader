DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  password VARCHAR NOT NULL,
  CONSTRAINT users_username_unique UNIQUE (username)
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
  CONSTRAINT assignments_title_unique UNIQUE (title)
);

INSERT INTO assignments (title, description, grader_url, container, part_id, files)
  VALUES (
    'Grader #1',
    'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
    'http://127.0.0.1:8021/api/v1/grader',
    'golangcourse_final',
    'HW1_game',
    '{"main.go"}'
  );
