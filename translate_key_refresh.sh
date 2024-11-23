CONFIG_FILE=".env"

update_trKey() {
    # Получаем новый токен
    trKey=$(yc iam create-token)
    bearer_token="Bearer $trKey"

    # Если trKey уже есть в .env, обновляем его
    if grep -q "^YANDEX_KEY=" "$CONFIG_FILE"; then
        sed -i "s/^YANDEX_KEY=.*/YANDEX_KEY=\"$bearer_token\"/" "$CONFIG_FILE"
    else
        # Если переменной trKey нет, добавляем её в конец файла
        echo "YANDEX_KEY=\"$bearer_token\"" >> "$CONFIG_FILE"
    fi

    echo "YANDEX_KEY обновлен: $bearer_token"
}

# Бесконечный цикл для обновления каждые 10 часов
while true; do
    update_trKey
    sleep 36000 # 10 часов в секундах
done
