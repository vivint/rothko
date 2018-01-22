port module Graph exposing (..)


port draw : () -> Cmd msg


port starting : (() -> msg) -> Sub msg


port done : (() -> msg) -> Sub msg
