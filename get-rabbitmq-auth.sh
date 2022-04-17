username="$(kubectl get secret rabbitmq-default-user -n rabbitmq -o jsonpath='{.data.username}' | base64 --decode)"
echo "username: $username"
password="$(kubectl get secret rabbitmq-default-user -n rabbitmq -o jsonpath='{.data.password}' | base64 --decode)"
echo "password: $password"
