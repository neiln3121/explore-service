-- +migrate Up

INSERT INTO decisions (recipient_id, actor_id, liked, updated_at)
VALUES ('user-1','user-2', true, now());
INSERT INTO decisions (recipient_id, actor_id, liked, updated_at)
VALUES ('user-1','user-3', false, now());
INSERT INTO decisions (recipient_id, actor_id, liked, updated_at)
VALUES ('user-1','user-4', true, now());
INSERT INTO decisions (recipient_id, actor_id, liked, updated_at)
VALUES ('user-1','user-5', true, now());

-- +migrate Down

DELETE FROM decisions WHERE recipient_id = 'user-1';
DELETE FROM decisions WHERE recipient_id = 'user-2';
DELETE FROM decisions WHERE recipient_id = 'user-3';
DELETE FROM decisions WHERE recipient_id = 'user-4';
DELETE FROM decisions WHERE recipient_id = 'user-5';