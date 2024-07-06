-- +goose Up
CREATE TABLE vacancies(
    id TEXT PRIMARY KEY,
    name TEXT,
    url TEXT,
    salary_from INTEGER,
    salary_to INTEGER,
    currency TEXT,
    experience INTEGER,
    employer_name TEXT,
    city_name TEXT
);



-- +goose Down
DROP TABLE vacancies;