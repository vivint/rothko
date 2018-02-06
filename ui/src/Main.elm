module Main exposing (..)

import Api
import Graph
import Html exposing (Html, text)
import Html.Attributes as Attr
import Html.Events as Ev
import Window
import Http
import Autocomplete as Auto
import Keyboard
import Task
import Time
import Process


-- MAIN


main =
    Html.programWithFlags
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Flags =
    {}



-- MSG


type Msg
    = NoOp
    | SetAutoState Auto.State
    | Select
    | DrawStarting
    | DrawDone Graph.Result
    | Render
    | RenderResponse (Result Http.Error String)
    | SetQueryText String
    | QueryTimer String
    | Query String
    | QueryResponse (Result Http.Error (List String))
    | KeyPress Keyboard.KeyCode



-- MODEL


type GraphState
    = Stale
    | Drawing
    | Fresh
    | Failed String


type QueryResult
    = QRNothing
    | QRWaiting
    | QRJust ( List String, Maybe Auto.State )


queryResult_new : List String -> QueryResult
queryResult_new results =
    QRJust ( results, Just Auto.new )


queryResult_mapAutoState : (Auto.State -> Auto.State) -> QueryResult -> QueryResult
queryResult_mapAutoState fn queryResult =
    case queryResult of
        QRJust ( results, Just state ) ->
            QRJust ( results, Just (fn state) )

        _ ->
            queryResult


queryResult_toMaybe : QueryResult -> Maybe ( List String, Maybe Auto.State )
queryResult_toMaybe queryResult =
    case queryResult of
        QRJust ( results, state ) ->
            Just ( results, state )

        _ ->
            Nothing


type alias Model =
    { graphState : GraphState
    , queryResult : QueryResult
    , queryText : String
    }



-- INIT


init : Flags -> ( Model, Cmd Msg )
init flags =
    ( { graphState = Stale
      , queryResult = QRNothing
      , queryText = ""
      }
    , Cmd.none
    )



-- UPDATE


withModel : Model -> Cmd Msg -> ( Model, Cmd Msg )
withModel model msg =
    ( model, msg )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        SetAutoState autoState ->
            ( { model
                | queryResult =
                    queryResult_mapAutoState
                        (always autoState)
                        model.queryResult
              }
            , Cmd.none
            )

        Select ->
            update_select model

        DrawStarting ->
            ( { model | graphState = Drawing }, Cmd.none )

        DrawDone result ->
            ( if result.ok then
                { model | graphState = Fresh }
              else
                { model | graphState = Failed result.error }
            , Cmd.none
            )

        Render ->
            Api.renderRequest model.queryText
                |> Api.render
                |> Http.send RenderResponse
                |> withModel model

        RenderResponse response ->
            case response of
                Ok result ->
                    ( model, Graph.draw result )

                Err error ->
                    Debug.log (toString error) ( model, Cmd.none )

        SetQueryText queryText ->
            case queryText of
                "" ->
                    ( { model | queryText = queryText, queryResult = QRNothing }
                    , Cmd.none
                    )

                _ ->
                    ( { model
                        | queryText = queryText
                        , queryResult = QRWaiting
                      }
                    , Process.sleep (250 * Time.millisecond)
                        |> Task.perform (always (QueryTimer queryText))
                    )

        QueryTimer queryText ->
            ( model
            , if queryText == model.queryText then
                utils_sendMessage <| Query queryText
              else
                Cmd.none
            )

        Query query ->
            Api.queryRequest query
                |> Api.query
                |> Http.send QueryResponse
                |> withModel model

        QueryResponse response ->
            ( case response of
                Ok results ->
                    { model | queryResult = queryResult_new results }

                _ ->
                    model
            , Cmd.none
            )

        KeyPress keyCode ->
            update_keyCode model keyCode


update_keyCode : Model -> Keyboard.KeyCode -> ( Model, Cmd Msg )
update_keyCode model keyCode =
    let
        length =
            model.queryResult
                |> queryResult_toMaybe
                |> Maybe.map Tuple.first
                |> Maybe.map List.length
                |> Maybe.withDefault 0
    in
        case keyCode of
            -- UP
            38 ->
                ( { model
                    | queryResult =
                        queryResult_mapAutoState
                            Auto.upPress
                            model.queryResult
                  }
                , Cmd.none
                )

            -- DOWN
            40 ->
                ( { model
                    | queryResult =
                        queryResult_mapAutoState
                            (Auto.downPress length)
                            model.queryResult
                  }
                , Cmd.none
                )

            -- ENTER
            13 ->
                update_select model

            -- ESCAPE
            27 ->
                ( { model | queryResult = QRNothing }, Cmd.none )

            _ ->
                ( model, Cmd.none )


update_select : Model -> ( Model, Cmd Msg )
update_select model =
    ( case model.queryResult of
        QRJust ( results, Just autoState ) ->
            let
                setSelection selected =
                    { model
                        | queryText = selected
                        , queryResult = QRNothing
                    }
            in
                Auto.selected autoState results
                    |> Maybe.map setSelection
                    |> Maybe.withDefault model

        _ ->
            model
    , utils_sendMessage Render
    )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Graph.starting DrawStarting
        , Graph.done DrawDone
        , Keyboard.downs KeyPress
        ]



-- VIEW


config : Auto.Config Msg
config =
    Auto.config
        { toMsg = SetAutoState
        , select = Select
        }


view : Model -> Html Msg
view model =
    Html.div []
        [ view_header
        , Html.br [] []
        , Html.div [ Attr.class "container" ]
            [ view_fullRow
                [ view_canvas
                ]
            , view_fullRow
                [ view_query model
                , view_errorText model
                ]
            ]
        ]


view_fullRow : List (Html msg) -> Html msg
view_fullRow children =
    Html.div [ Attr.class "row" ]
        [ Html.div [ Attr.class "col-sm-12" ] children
        ]


view_centeredRow : List (Html msg) -> Html msg
view_centeredRow children =
    Html.div [ Attr.class "row" ]
        [ Html.div [ Attr.class "col-sm-8 col-sm-offset-2" ] children
        ]


view_canvas : Html msg
view_canvas =
    Html.div [ Attr.id "canvas-container" ]
        [ Html.canvas [ Attr.id "canvas" ] [] ]


view_errorText : Model -> Html msg
view_errorText model =
    case model.graphState of
        Failed error ->
            Html.pre [] [ text error ]

        _ ->
            text ""


view_header : Html Msg
view_header =
    Html.header [ Attr.class "sticky" ]
        [ Html.a [ Attr.href "#", Attr.class "logo" ] [ text "Rothko" ]
        ]


view_query : Model -> Html Msg
view_query model =
    let
        attrs =
            [ Attr.value model.queryText
            , Attr.id "query"
            , Attr.style [ ( "width", "100%" ) ]
            , Ev.onInput SetQueryText
            ]

        autocomplete =
            case model.queryResult of
                QRJust ( results, Just autoState ) ->
                    Just (Auto.view config autoState results)

                _ ->
                    Nothing
    in
        []
            |> utils_add (Html.input attrs [])
            |> utils_maybeAdd autocomplete
            |> view_fullRow



-- UTILS


utils_add : a -> List a -> List a
utils_add val list =
    list ++ [ val ]


utils_maybeAdd : Maybe a -> List a -> List a
utils_maybeAdd val list =
    case val of
        Nothing ->
            list

        Just val ->
            utils_add val list


utils_when : Bool -> a -> Maybe a
utils_when cond val =
    if cond then
        Just val
    else
        Nothing


utils_sendMessage : msg -> Cmd msg
utils_sendMessage m =
    Task.perform (\_ -> m) (Task.succeed ())


utils_titleCase : String -> String
utils_titleCase val =
    let
        upperWord word =
            String.toUpper (String.left 1 word) ++ (String.dropLeft 1 word)
    in
        val
            |> String.words
            |> List.map upperWord
            |> String.join " "
