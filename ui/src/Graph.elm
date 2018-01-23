port module Graph
    exposing
        ( Result
        , draw
        , starting
        , done
        )

{-|
  since some ports have arguments that we are ignoring, we define them with a
  P suffix, and rexport them with the arguments bound.
-}


type alias Result =
    { ok : Bool
    , error : String
    }


port drawP : () -> Cmd msg


draw : Cmd msg
draw =
    drawP ()


port startingP : (() -> msg) -> Sub msg


starting : msg -> Sub msg
starting msg =
    startingP <| always msg


port done : (Result -> msg) -> Sub msg
