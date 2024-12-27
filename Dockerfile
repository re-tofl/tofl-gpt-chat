FROM postgres:16

# Установка необходимых инструментов
RUN apt-get update && apt-get install -y \
    postgresql-server-dev-16 \
    gcc \
    make \
    git \
    && git clone https://github.com/pgvector/pgvector.git \
    && cd pgvector && make && make install \
    && apt-get remove -y gcc make git && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* pgvector

# Копирование вашего SQL файла
COPY init.sql /docker-entrypoint-initdb.d/

# Оставляем стандартный entrypoint
EXPOSE 5432
CMD ["postgres"]
