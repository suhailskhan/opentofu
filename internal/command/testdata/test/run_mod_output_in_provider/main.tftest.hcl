variables {
  username = "test_user"
}

provider "test" {
  username = var.username
  password = run.setup.password
  data_prefix = "test"
  resource_prefix = "test"
}

run "setup" {
  module {
    source = "./first"
  }
}

run "validate" {}
