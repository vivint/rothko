module Autocomplete
    exposing
        ( State
        , new
        , downPress
        , upPress
        , selected
        , Config
        , config
        , view
        )

import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events as Ev
import Json.Decode as Json


-- STATE


type State
    = State
        { highlight : Maybe Int
        }


new : State
new =
    State { highlight = Nothing }


downPress : Int -> State -> State
downPress largest (State state) =
    let
        update highlight =
            let
                newIndex =
                    Maybe.map ((+) 1) highlight |> Maybe.withDefault 0
            in
                when (newIndex < largest) newIndex |> orElse highlight
    in
        State { state | highlight = update state.highlight }


upPress : State -> State
upPress (State state) =
    let
        update highlight =
            let
                newIndex =
                    Maybe.map ((flip (-)) 1) highlight |> Maybe.withDefault -1
            in
                when (newIndex >= 0) newIndex |> orElse highlight
    in
        State { state | highlight = update state.highlight }


selected : State -> List String -> Maybe String
selected (State state) completions =
    state.highlight
        |> Maybe.map (\n -> List.drop n completions)
        |> Maybe.andThen List.head



-- CONFIG


type Config msg
    = Config
        { toMsg : State -> msg
        , select : msg
        }


config :
    { toMsg : State -> msg
    , select : msg
    }
    -> Config msg
config { toMsg, select } =
    Config
        { toMsg = toMsg
        , select = select
        }



-- VIEW


view : Config msg -> State -> List String -> Html msg
view config state completions =
    let
        attrs =
            []
                |> add (Attr.class "ui-menu")
    in
        Html.ul attrs (viewCompletions config state completions)


viewCompletions : Config msg -> State -> List String -> List (Html msg)
viewCompletions config state completions =
    List.indexedMap (viewCompletion config state (List.length completions)) completions


viewCompletion : Config msg -> State -> Int -> Int -> String -> Html msg
viewCompletion config (State state) largest index completion =
    let
        active =
            state.highlight == Just index

        attrs =
            []
                |> add (Attr.class "ui-menu-item")
                |> maybeAdd (when active (Attr.class "ui-menu-item-active"))
                |> add (mouseEnter config (State state) index)
                |> add (mouseLeave config (State state) index)
                |> add (onClick config)
    in
        Html.li
            attrs
            [ Html.div
                [ Attr.class "ui-menu-item-wrapper" ]
                [ Html.text completion ]
            ]


mouseEnter : Config msg -> State -> Int -> Html.Attribute msg
mouseEnter (Config config) (State state) index =
    Ev.onMouseOver <| config.toMsg (State { state | highlight = Just index })


mouseLeave : Config msg -> State -> Int -> Html.Attribute msg
mouseLeave (Config config) (State state) index =
    let
        newState =
            if state.highlight == Just index then
                { state | highlight = Nothing }
            else
                state
    in
        Ev.onMouseLeave <| config.toMsg (State newState)


onClick : Config msg -> Html.Attribute msg
onClick (Config config) =
    Ev.onClick config.select



-- UTILS


add : a -> List a -> List a
add val list =
    list ++ [ val ]


maybeAdd : Maybe a -> List a -> List a
maybeAdd val list =
    case val of
        Nothing ->
            list

        Just val ->
            add val list


when : Bool -> a -> Maybe a
when cond val =
    if cond then
        Just val
    else
        Nothing


orElse : Maybe a -> Maybe a -> Maybe a
orElse otherwise val =
    case val of
        Just _ ->
            val

        Nothing ->
            otherwise
