module Main exposing (..)

import Graph
import Html exposing (text)


main =
    Html.programWithFlags
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Flags =
    {}


type Msg
    = Starting
    | Done


init : Flags -> ( (), Cmd b )
init flags =
    () ! [ Graph.draw () ]


update msg model =
    model ! []


subscriptions model =
    Sub.batch
        [ Graph.starting <| const Starting
        , Graph.done <| const Done
        ]


view model =
    text "Hello World"


const x y =
    y
