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
	* [vouquet/florist](https://github.com/vouquet/florist) へアクセスできない場合、"github.com/vouquet/florist"を[Floristライブラリ要求仕様](#florist-ライブラリ要求仕様)に従って作成したライブラリへ書き換える
2. build用コンテナの用意と各種ライブラリの解決
	1. `cd <repository>/docker`
		* proxy 配下の場合は、`<repository>/docker/dev-go-vouquet/Dockerfile` の、`http_proxy`及び`https_proxy` を書き換える
	2. `make run`
	3. `make godep`
3. build
	1. `make gobuild`

## Florist ライブラリ要求仕様

1. パッケージ名が、florist(`source L1: package florist`) であること
2. パッケージ内に、グローバル変数 `MEMBERS map[string]func() base.Florist` を有し、全てのflorist名とNew関数がmapに格納されていること
	* `base.Florist` のインタフェースに合致しないものはbuildで弾かれます
* [サンプル](../../sample/lib/florist/florist.go)

#### base.Florist の対応説明
* 説明
	* `base.Florist` インタフェースにハマる構造体を定義することで、Vouquetのロジックとして利用できます
		* `base.FloristBase` を組み込みインポートすることで、大体の関数は用意されます
			* 組み込みインポートをするので、個別のバッファやフラグを用意する場合、子構造体として格納することをお勧めします
		* 開発者が個別に定義する必要があるものは以下の2つのみです
			* `Init([]*farm.State) error`
				* 初期データの生成や計算を想定しています
				* `func main()`のgoroutineのスレッドで動きます
				* 起動直後に呼ばれます
				* 起動地点から、1時間前のデータが時間昇順で渡されます
			* `Action(*farm.State) error`
				* 実際の取引を行う内容が作成されることを想定しています
				* `func main()`以外のgoroutineのスレッドで動きます
				* メインではない、goroutineのスレッドで動きます
				* 概ね1秒に1件の最新のデータを渡し、関数が呼ばれます
					* 取引所が停止した場合、停止時間分のデータが受信できない（例えばいきなり10時間後のデータが渡されるなど）ことが想定されます
					* 計算前には、取得時間の確認をお勧めします
* 取引を行う際に利用する関数
	* `<your florist>.Planter()`で取得できる、`farm.Planter`。取引に利用するインタフェース
		* `farm.Planter.SetSeed(o_type string, size float64, opt *OpeOption) error`
			* 引数 o_type では、売り/買いを指定します。パラメータは、`farm.TYPE_BUY`, `farm.TYPE_SELL`を指定ください
			* 引数 size では、取引のサイズを渡します。`vouquet.Florist.Run`の、`size`を渡すことを想定しています
			* 引数 opt では、実行のオプションを指定します
		* `farm.Planter.ShowSproutList() ([]*farm.Sprout, error)`
		* `farm.Planter.Harvest(sp *Sprout, opt *OpeOption)`
			* 引数 sp では、利確する、`farm.Sprout`を 渡してください
			* 引数 opt では、実行のオプションを指定します
	* `<your florist>.Size()`で取得できる、`サイズ(float64)`
		* バイナリの実行時に、引数で指定するサイズが渡されるようになります
		* `farm.Planter.SetSeed`をする際の`size` で活用することを想定しています
			* `size`を任意のタイミングで分割したいケースを想定し、`size`は利用者側が指定できるようにしています
	* `farm.OpeOption` に存在する要素
		* `Price`
			* 実行時の値段を指定します。`Stream: true`時は無視されます
		* `Stream`
			* 実行時に指値か、成行を指定します。`Stream: true`で成行です
			* `v0.0.3`現在、指値未対応で、全ての注文が成行です
				* 進捗については、[https://github.com/vouquet/vouquet/issues/38](https://github.com/vouquet/vouquet/issues/38) をご覧ください
* その他
	* `<your florist>.Logger()`で取得できる、`base.Logger`
		* 並列動作で呼ばれることや、上段のWrapperが拡張されうるので、`import log`のようなログの出力は非推奨です。本構造体を利用ください
		* 利用できる出力は以下です
			* `WriteErr(string, ...interface{})`
				* stderrに出力します
			* `WriteMsg(string, ...interface{})`
				* stdoutに出力します
				* `vqt_eval`では、`-v`で出力されます
			* `WriteDebug(string, ...interface{})`
				* stdoutに出力します
				* `vqt_eval`では、`-vv`で出力されます
