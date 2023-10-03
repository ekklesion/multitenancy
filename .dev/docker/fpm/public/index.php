<?php

$data = [
    'secret' => getenv('APP_SECRET'),
    'database' => getenv('DATABASE_URL'),
    'mailer' => getenv('MAILER_DSN'),
];

header('Content-Type: application/json', true, 200);

echo json_encode($data);