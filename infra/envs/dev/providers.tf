# AWS Provider + default_tags (Q-I11)
#
# region: locals.region (= "ap-northeast-1") と二重定義に見えるが、
# provider の `region` は data source / variable の参照ができない
# (provider 初期化が module 評価より前のため)。よって provider 側は
# ハードコードのままとし、locals.region は module への変数注入だけに使う。

provider "aws" {
  region = "ap-northeast-1"

  # 注意: Unit タグは default_tags に入れない。
  # 後続 Unit B/C/D/E のリソースが同じ envs/dev に追加された際、
  # provider レベルの default が override されないと「Unit=auth」が
  # 全リソースに伝播してコスト集計が壊れるため、Unit タグは
  # 各 module 内で個別に `tags = merge({Unit = "..."}, ...)` で付ける。
  default_tags {
    tags = {
      Project   = "goro2pay"
      Env       = "dev"
      ManagedBy = "terraform"
    }
  }
}
