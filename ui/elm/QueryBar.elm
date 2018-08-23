module QueryBar exposing
    ( Config
    , Model
    , Msg
    , new
    , subscriptions
    , update
    , view
    )

import Api
import Browser.Events as Bev
import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events as Ev
import Http
import Json.Decode as Decode
import Process
import Task
import Time



-- MODEL


type Query
    = None
    | Got (Maybe Int) (List String)


type Model
    = Model
        { value : String
        , query : Query
        }


new : String -> Model
new value =
    Model
        { value = value
        , query = None
        }



-- MSG


type Msg
    = SetValue String
    | Focus
    | Timer String
    | QueryResponse String (Result Http.Error (List String))
    | Dismiss
    | Select String
    | KeyDown String
    | MouseOver Int
    | MouseLeave Int



-- CONFIG


type alias Config model msg =
    { get : model -> Model
    , set : model -> Model -> model
    , wrap : Msg -> msg
    , select : String -> msg
    }



-- SUBSCRIPTIONS


subscriptions : Config model msg -> model -> Sub msg
subscriptions config model =
    doSubscriptions (config.get model)
        |> Sub.map config.wrap


doSubscriptions : Model -> Sub Msg
doSubscriptions (Model model) =
    let
        decoder =
            Decode.string
                |> Decode.field "key"
                |> Decode.map KeyDown
    in
    Bev.onKeyDown decoder



-- UPDATE


update : Config model msg -> model -> Msg -> ( model, Cmd msg )
update config model msg =
    let
        ( newModel, cmd, parentCmd ) =
            doUpdate config (config.get model) msg
    in
    ( config.set model newModel
    , Cmd.batch
        [ Cmd.map config.wrap cmd
        , parentCmd
        ]
    )


doUpdate : Config model msg -> Model -> Msg -> ( Model, Cmd Msg, Cmd msg )
doUpdate config (Model model) msg =
    case msg of
        SetValue value ->
            ( Model { model | value = value, query = None }
            , Process.sleep 250
                |> Task.perform (always (Timer value))
            , Cmd.none
            )

        Focus ->
            ( Model model
            , Api.queryRequest model.value
                |> Api.doQuery
                |> Http.send (QueryResponse model.value)
            , Cmd.none
            )

        Timer value ->
            ( Model model
            , Api.queryRequest value
                |> Api.doQuery
                |> Http.send (QueryResponse value)
                |> when (value == model.value)
                |> Maybe.withDefault Cmd.none
            , Cmd.none
            )

        QueryResponse value response ->
            ( Result.toMaybe response
                |> Maybe.andThen (when (value == model.value))
                |> Maybe.map (\r -> { model | query = Got Nothing r })
                |> Maybe.withDefault model
                |> Model
            , Cmd.none
            , Cmd.none
            )

        Dismiss ->
            ( Model { model | query = None }
            , Cmd.none
            , Cmd.none
            )

        Select value ->
            ( Model { model | value = value, query = None }
            , Cmd.none
            , sendMessage (config.select value)
            )

        KeyDown keyCode ->
            case ( keyCode, model.query ) of
                ( _, Got highlight completions ) ->
                    let
                        ( newHighlight, cmd ) =
                            updateKeyCode highlight completions keyCode
                    in
                    ( Model { model | query = Got newHighlight completions }
                    , cmd
                    , Cmd.none
                    )

                -- Enter with no completions triggers a redraw
                ( "Enter", _ ) ->
                    ( Model model
                    , Cmd.none
                    , sendMessage (config.select model.value)
                    )

                _ ->
                    ( Model model
                    , Cmd.none
                    , Cmd.none
                    )

        MouseOver index ->
            case model.query of
                Got _ completions ->
                    ( Model { model | query = Got (Just index) completions }
                    , Cmd.none
                    , Cmd.none
                    )

                _ ->
                    ( Model model
                    , Cmd.none
                    , Cmd.none
                    )

        MouseLeave index ->
            case model.query of
                Got (Just current) completions ->
                    if current == index then
                        ( Model { model | query = Got Nothing completions }
                        , Cmd.none
                        , Cmd.none
                        )

                    else
                        ( Model model
                        , Cmd.none
                        , Cmd.none
                        )

                _ ->
                    ( Model model
                    , Cmd.none
                    , Cmd.none
                    )


updateKeyCode : Maybe Int -> List String -> String -> ( Maybe Int, Cmd Msg )
updateKeyCode highlight completions keyCode =
    case keyCode of
        "ArrowUp" ->
            case highlight of
                Nothing ->
                    ( Nothing, Cmd.none )

                Just val ->
                    ( Just <| max (val - 1) 0, Cmd.none )

        "ArrowDown" ->
            let
                length =
                    List.length completions
            in
            case ( length, highlight |> Maybe.withDefault -1 ) of
                ( 0, _ ) ->
                    ( Nothing, Cmd.none )

                ( _, val ) ->
                    ( Just <| min (val + 1) (length - 1), Cmd.none )

        "Enter" ->
            let
                select index =
                    (List.drop index >> List.head) completions
            in
            case highlight |> Maybe.andThen select of
                Just selected ->
                    ( highlight, sendMessage <| Select selected )

                Nothing ->
                    ( highlight, Cmd.none )

        "Escape" ->
            ( highlight, sendMessage Dismiss )

        _ ->
            ( highlight, Cmd.none )



-- VIEW


view : Config model msg -> Model -> Html msg
view config model =
    doView model
        |> Html.map config.wrap


doView : Model -> Html Msg
doView (Model model) =
    let
        input =
            [ Html.input
                [ Attr.value model.value
                , Attr.style "width" "100%"
                , Attr.type_ "text"
                , Attr.spellcheck False
                , Ev.onInput SetValue
                , Ev.onFocus Focus
                , Ev.onBlur Dismiss
                ]
                []
            ]

        autocomplete =
            case model.query of
                Got highlight completions ->
                    [ viewAutocomplete highlight completions ]

                _ ->
                    []

        children =
            []
                ++ input
                ++ autocomplete
    in
    Html.div [] children


viewAutocomplete : Maybe Int -> List String -> Html Msg
viewAutocomplete highlight completions =
    case completions of
        [] ->
            Html.text ""

        _ ->
            Html.ul
                [ Attr.class "auto-menu" ]
                (viewCompletions highlight completions)


viewCompletions : Maybe Int -> List String -> List (Html Msg)
viewCompletions highlight completions =
    List.indexedMap (viewCompletion highlight) completions


viewCompletion : Maybe Int -> Int -> String -> Html Msg
viewCompletion highlight index completion =
    let
        activeAttrs =
            if Just index == highlight then
                [ Attr.class "auto-menu-item-active" ]

            else
                []

        attrs =
            [ Attr.class "auto-menu-item"
            , Ev.onMouseOver (MouseOver index)
            , Ev.onMouseLeave (MouseLeave index)
            , Ev.onMouseDown (Select completion)
            ]
                ++ activeAttrs
    in
    Html.li attrs
        [ Html.div
            [ Attr.class "auto-menu-item-wrapper" ]
            [ Html.text completion ]
        ]



-- UTILS


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())


when : Bool -> a -> Maybe a
when cond val =
    if cond then
        Just val

    else
        Nothing
