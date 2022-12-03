import gleam/bit_builder
import gleam/erlang/process
import gleam/otp/actor
import gleam/result
import gleam/dynamic
import gleam/json
import gleam/float
import gleam/int
import gleam/list
import gleam/string_builder
import glisten/acceptor
import glisten/handler
import glisten/socket
import glisten/tcp
import glisten

pub fn main() {
  // TODO: msgs are not sent on newline bounds
  handler.func(fn(msg, state) {
    prime(msg, state.socket)
    actor.Continue(state)
  })
  |> acceptor.new_pool
  |> glisten.serve(1337, _)
  |> result.map(fn(_) { process.sleep_forever() })
}

type Request {
  Request(method: String, number: Float)
}

fn prime(msg: BitString, socket: socket.Socket) {
  let decoder =
    dynamic.decode2(
      Request,
      dynamic.field("method", of: dynamic.string),
      dynamic.field("number", of: dynamic.any(of: [
          dynamic.float,
          fn(x) {
            dynamic.int(x)
            |> result.map(int.to_float)
          },
        ]),
      ),
    )

  let req = json.decode_bits(from: msg, using: decoder)

  let resp =
    case req {
      Error(_) -> json.object([#("error", json.string("invalid request"))])
      Ok(Request(method: "isPrime", number: number)) ->
        case float.floor(number) == number && number >. 0. {
          True ->
            json.object([
              #("method", json.string("isPrime")),
              #("number", json.float(number)),
              #("prime", json.bool(is_prime(float.truncate(number)))),
            ])
          False ->
            json.object([
              #("method", json.string("isPrime")),
              #("number", json.float(number)),
              #("prime", json.bool(False)),
            ])
        }
      _ -> json.object([#("error", json.string("unrecognized method"))])
    }
    |> json.to_string_builder
    |> string_builder.append("\n")
    |> bit_builder.from_string_builder

  let _ = tcp.send(socket, resp)
}

fn is_prime(number: Int) -> Bool {
  case number {
    0 | 1 -> False
    2 | 3 | 5 | 7 -> True
    _ -> {
      let max =
        number
        |> int.square_root
        |> result.map(float.round)
        |> result.unwrap(or: 3)
      int.is_odd(number) && list.range(3, max)
      |> list.filter(int.is_odd)
      |> list.all(fn(x) { number % x != 0 })
    }
  }
}
