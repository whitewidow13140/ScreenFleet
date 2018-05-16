
CREATE DATABASE tvserver;
CREATE USER zgegos
WITH PASSWORD 'bite';
GRANT ALL ON DATABASE tvserver TO tvserver;
CREATE TABLE television
(
    id varchar
     (20),
    name varchar
     (30),
    ip varchar(30),
    status int,
    composition_id varchar(50)
);
INSERT INTO television
    (id, name, ip, status, composition_id)
VALUES
    ('00000000001', 'Television-01', '192.168.0.1', 4, '5');
INSERT INTO television
    (id, name, ip, status, composition_id)
VALUES
    ('00000000002', 'Television-02', '192.168.0.2', 4, '5');
INSERT INTO television
    (id, name, ip, status, composition_id)
VALUES
    ('00000000003', 'Television-03', '192.168.0.3', 4, '5');

