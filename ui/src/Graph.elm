port module Graph
    exposing
        ( Result
        , draw
        , starting
        , done
        )


type alias Result =
    { ok : Bool
    , metric : String
    , error : String
    }


port draw : String -> Cmd msg


port starting : (String -> msg) -> Sub msg


port done : (Result -> msg) -> Sub msg
