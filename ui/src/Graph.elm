port module Graph
    exposing
        ( Result
        , draw
        , starting
        , done
        )


type alias Result =
    { ok : Bool
    , error : String
    }


port draw : String -> Cmd msg


port startingP : (() -> msg) -> Sub msg


starting : msg -> Sub msg
starting msg =
    startingP <| always msg


port done : (Result -> msg) -> Sub msg
