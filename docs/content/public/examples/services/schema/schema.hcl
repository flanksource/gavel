variable "greeting" {
  type = string
}

schema "public" {}

table "messages" {
  schema = schema.public
  column "message" {
    type = text
    default = var.greeting
  }
}

table "instance_events" {
  schema = schema.public
  column "database_name" {
    type = text
  }
}
