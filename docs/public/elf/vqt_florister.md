vqt_florister
===
floristをimportし、rateを元に取引をおこないます

## img

![img](../media/vqt_florister.png)

## usage
```
vqt_florister [-c <config path>] <NAMEofFlorist> <SEED> <SOIL> <SIZE>
```

* `-c <config path>`
	* [config](../../../var.in/service/vouquet/etc/vouquet.conf) を指定します
* `<NameofFlorist>`
	* 使用するFlorist名を指定します
* `<SEED>`
	* 仮想通貨名を指定します
* `<SOIL>`
	* 仮想通貨取引所を指定します
* `<SIZE>`
	* 取引を行うサイズを指定します
	* `0.01BTC`単位で取引を行う場合の指定は、`0.01`です

