locals {
  # Unit A 命名規則 gp-{env}-{resource} (Q-I4) に整合
  table_name  = "gp-${var.env}-suggestion"
  policy_name = "gp-${var.env}-suggestion-policy"
}
