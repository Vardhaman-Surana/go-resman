FROM heroku/heroku:18
COPY ./bin/server /
COPY ./database/1_data_load.up.sql /database/
EXPOSE 4000
CMD ["/server"]
