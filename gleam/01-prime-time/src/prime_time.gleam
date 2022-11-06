import gleam/bit_builder
import gleam/erlang/process
import gleam/otp/actor
import gleam/result
import gleam/io
import gleam/dynamic
import gleam/json
import gleam/float
import gleam/int
import gleam/list

import glisten/acceptor
import glisten/handler
import glisten/socket
import glisten/tcp
import glisten

pub fn main() {
  handler.func(fn(msg, state) {
    prime(msg, state.socket)
    actor.Continue(state)
  })
  |> acceptor.new_pool
  |> glisten.serve(1337, _)
  |> result.map(fn(_) { process.sleep_forever() })
}

type Request {
  FloatRequest(method: String, number: Float)
  IntRequest(method: String, number: Int)
}

fn prime(msg: BitString, socket: socket.Socket) -> Result(Nil, json.DecodeError) {
  let float_decoder = dynamic.decode2(
    FloatRequest,
    dynamic.field("method", of: dynamic.string),
    dynamic.field("number", of: dynamic.float),
  )
  let int_decoder = dynamic.decode2(
    IntRequest,
    dynamic.field("method", of: dynamic.string),
    dynamic.field("number", of: dynamic.int),
  )

  let req = json.decode_bits(from: msg, using: int_decoder)
    |> result.or(json.decode_bits(from: msg, using: float_decoder))


  let resp = case req {
    Error(_) -> json.object([
      #("error", json.string("invalid request")),
    ])
    Ok(IntRequest(method: "isPrime", number: number)) if number < 0 -> json.object([
      #("method", json.string("isPrime")),
      #("number", json.int(number)),
      #("prime", json.bool(False)),
    ])
    Ok(IntRequest(method: "isPrime", number: number)) -> json.object([
      #("method", json.string("isPrime")),
      #("number", json.int(number)),
      #("prime", json.bool(is_prime(number))),
    ])
    Ok(FloatRequest(method: "isPrime", number: number)) -> {
      case float.floor(number) == number {
        True -> json.object([
          #("method", json.string("isPrime")),
          #("number", json.float(number)),
          #("prime", json.bool(is_prime(float.truncate(number)))),
        ])
        False -> json.object([
          #("method", json.string("isPrime")),
          #("number", json.float(number)),
          #("prime", json.bool(False)),
        ])
      }
    }
    _ -> {todo}
  }
  tcp.send(socket, resp
    |> json.to_string_builder
    |> bit_builder.from_string_builder
  )
  tcp.send(socket, bit_builder.from_string("\n"))

  Ok(Nil)
}

fn is_prime(number: Int) -> Bool {
  case number {
    0 | 1 -> False
    2 | 3 | 5 | 7 -> True
    _ -> {
      let max = number
      |> int.square_root
      |> result.unwrap(0.0)
      |> float.round
      case int.is_odd(number) {
        True -> list.range(3, max)
          |> list.filter(int.is_odd)
          |> list.all(fn(x) { number % x != 0 })
        False -> False
      }
    }
  }
}
