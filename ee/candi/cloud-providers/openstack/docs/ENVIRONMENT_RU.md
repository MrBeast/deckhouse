---
title: "Cloud provider — OpenStack: подготовка окружения"
---

Для того чтобы Deckhouse мог управлять ресурсами в облаке OpenStack, ему необходимо подключиться к OpenStack API. 
Доступы пользователя, необходимые для подключения к OpenStack API, находятся в openrc-файле (OpenStack RC file).

Информация о получении openrc-файла с помощью стандартного веб-интерфейса OpenStack и о способах его использования доступна в [документации OpenStack](https://docs.openstack.org/ocata/admin-guide/common/cli-set-environment-variables-using-openstack-rc.html#download-and-source-the-openstack-rc-file).

Если вы используете OpenStack API cloud-провайдера, то интерфейс получения openrc-файла может быть другим. 
