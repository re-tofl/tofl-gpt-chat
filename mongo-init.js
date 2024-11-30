use admin
db.createUser({
    user: "tofl_user",
    pwd: "tofl_password",
    roles: [
        {
            role: "readWrite",
            db: "tofl_gpt_chat"
        }
    ]
});
