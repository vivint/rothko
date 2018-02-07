module QueryBar
    exposing
        ( Model
        , new
        , Msg
        , Config
        , subscriptions
        , update
        , view
        )

import Api
import Task
import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events as Ev
import Keyboard
import Process
import Time
import Http


-- MODEL


type Query
    = None
    | Got (Maybe Int) (List String)


queryToMaybe : Query -> Maybe ( Maybe Int, List String )
queryToMaybe query =
    case query of
        None ->
            Nothing

        Got highlight completions ->
            Just ( highlight, completions )


type Model
    = Model
        { value : String
        , query : Query
        }


new : Model
new =
    Model
        { value = ""
        , query = None
        }



-- MSG


type Msg
    = SetValue String
    | Timer String
    | QueryResponse String (Result Http.Error (List String))
    | Dismiss
    | Select String
    | KeyDown Keyboard.KeyCode
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
    Keyboard.downs KeyDown



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
            , Process.sleep (250 * Time.millisecond)
                |> Task.perform (always (Timer value))
                |> when (value /= "")
                |> Maybe.withDefault Cmd.none
            , Cmd.none
            )

        Timer value ->
            ( Model model
            , Api.queryRequest value
                |> Api.query
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
                ( 13, _ ) ->
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


updateKeyCode : Maybe Int -> List String -> Keyboard.KeyCode -> ( Maybe Int, Cmd Msg )
updateKeyCode highlight completions keyCode =
    case keyCode of
        -- UP
        38 ->
            case highlight of
                Nothing ->
                    ( Nothing, Cmd.none )

                Just val ->
                    ( Just <| max (val - 1) 0, Cmd.none )

        -- DOWN
        40 ->
            let
                length =
                    List.length completions
            in
                case ( length, highlight |> Maybe.withDefault -1 ) of
                    ( 0, _ ) ->
                        ( Nothing, Cmd.none )

                    ( _, val ) ->
                        ( Just <| min (val + 1) (length - 1), Cmd.none )

        -- ENTER
        13 ->
            let
                select completions index =
                    (List.drop index >> List.head) completions
            in
                case highlight |> Maybe.andThen (select completions) of
                    Just selected ->
                        ( highlight, sendMessage <| Select selected )

                    Nothing ->
                        ( highlight, Cmd.none )

        -- ESCAPE
        27 ->
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
                , Attr.style [ ( "width", "100%" ) ]
                , Ev.onInput SetValue
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
    Html.ul [ Attr.class "auto-menu" ] (viewCompletions highlight completions)


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
            , Ev.onClick (Select completion)
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
