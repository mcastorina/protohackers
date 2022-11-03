import gleam/bit_builder
import gleam/erlang/process
import gleam/otp/actor
import gleam/result
import glisten/acceptor
import glisten/handler
import glisten/tcp
import glisten

pub fn main() {
  handler.func(fn(msg, state) {
    assert Ok(_) = tcp.send(state.socket, bit_builder.from_bit_string(msg))
    actor.Continue(state)
  })
  |> acceptor.new_pool
  |> glisten.serve(1337, _)
  |> result.map(fn(_) { process.sleep_forever() })
}
