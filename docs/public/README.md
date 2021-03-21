Vouquet 公開仕様
===

![Dataflow](./media/Vouquet_Dataflow.png)

## 各種実行ファイルの説明

* [vqt_registrar](./elf/vqt_registrar.md)
	* rateをmysqlへ記録します
* [vqt_florister](./elf/vqt_florister.md)
	* floristをimportし、rateを元に取引をおこないます
* [vqt_eval](./elf/vqt_eval.md)
	* floristをimportし、過去のrateを流し込むことで、擬似的に過去の取引結果を再現します
	* floristの性能評価をおこないます
* [vqt_noticer](./elf/vqt_noticer.md)
	* 取引所の結果を取得し、通知します
	* 現在は、Twitterのみ通知可能です

## Florist 開発/リリース 方法

1. Floristの用意
	* [vouquet/florist](https://github.com/vouquet/florist) へアクセスできない場合、"github.com/vouquet/florist"を[Floristライブラリ要求仕様]()に従って作成したライブラリへ書き換える
2. build用コンテナの用意と各種ライブラリの解決
	1. `cd <repository>/docker`
		* proxy 配下の場合は、`<repository>/docker/dev-go-vouquet/Dockerfile` の、`http_proxy`及び`https_proxy` を書き換える
	2. `make run`
	3. `make godep`
3. build
	1. `make gobuild`

## Florist ライブラリ要求仕様

1. パッケージ名が、florist(`source L1: package florist`) であること
2. 定数 `florist.MEMBERS []string` を有し、全てのflorist名がsliceに格納されていること
3. 関数 `NewFlorist(name string, p farm.Planter, init_status []*farm.State, log logger) (vouquet.Florist, error)` を有すること
	* 引数
		1. `name` は、定数florist.MEMBERS に含まれる1つを渡します
		2. `planter` は、任意のfarm.Planter を渡します
		3. `init_status` は、過去1時間の状態を渡します。過去データを使い判断軸を生成する場合利用できます
		4. `log` は、farm.loggerが渡されます。各種実行ファイル上で安全に出力する為の構造体です。出力を行う場合はこちらを使うことを推奨します
	* 動作
		* `name`でswitch等をおこない、任意のvouquet.Floristに対応するfloristを返却してください
4. vouquet.Florist への対応
	* 説明
		* `vouquet.Florist`では、`Run(context.Context, float64, <- chan *farm.State) error`を有すること
			* 引数
				* `context.Context`
					* 停止時の`context.Context.Done()` を共有します
				* `float64`
					* 取引のサイズを共有します。1BTC毎に取引をする想定の場合、1を共有します
				* `<- chan *farm.State`
					* 最短1秒ごとに現在の`farm.State`を共有するchannelを共有します
	* 動作要求仕様
		* `chan *farm.State`から、現在の値を受け取り、`farm.Planter`に対して取引を行うこと
		* 動作フロー概要
			1. `chan *farm.State`から`farm.State`を取得
			2. `farm.Planter.SetSeed`によって、ポジションの作成
			3. `farm.Planter.ShowSproutList`によって、作成したポジション一覧の確認
			4. `farm.Planter.Harvest`によって、ポジションの利確を行う
		* `farm.Planter`で取引に利用する代表的なインタフェース
			* `farm.Planter.SetSeed(seed_name string, size float64, price float64) error`
				* 引数 seed_name では、仮想通貨名を渡します
				* 引数 size では、取引のサイズを渡します。`vouquet.Florist.Run`の、`size`を渡すことを想定しています
				* 引数 price では、取引の値段を渡します
					* `v0.0.2`現在、指値未対応で、全ての注文が成行です。この値はシミュレーション(`vqt_eval`)時のみ活用します
			* `farm.Planter.ShowSproutList() ([]*farm.Sprout, error)`
			* `farm.Planter.Harvest(sp *Sprout, price float64)`
				* 引数 sp では、利確する、`farm.Sprout`を 渡してください
				* 引数 price では、取引の値段を渡します
					* `v0.0.2`現在、指値未対応で、全ての注文が成行です。この値はシミュレーション(`vqt_eval`)時のみ活用します

