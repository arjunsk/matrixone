CREATE TABLE shirts (
                        name VARCHAR(40),
                        size ENUM('x-small', 'small', 'medium', 'large', 'x-large')
);
INSERT INTO shirts (name, size) VALUES ('dress shirt','large'), ('t-shirt','medium'), ('polo shirt','small');
SELECT name, size FROM shirts;
SELECT name, size FROM shirts WHERE size = 'medium';
DELETE FROM shirts where size = 'large';
SELECT name, size FROM shirts;
DROP TABLE shirts;

create table t_enum(a int,b enum('1','2','3','4','5'));
insert into t_enum values(1,1);
select * from t_enum;
drop table t_enum;
